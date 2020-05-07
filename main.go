/*
Программа генерирует лог в директории ./logs/logrus.log
через неравные промежутки времени.
Каждая строка имеет вид:
{"level":"info","msg":"Рейс задерживается на 1.449s","status":"INFO","time":"2020-04-22T15:21:07+03:00","wait":1449}

*/

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-extras/elogrus.v7"
)

// имя файла лога
var logFileName = "logrus.log"

// максимальное число записей в логе
var maxRecords int64 = 10

// максимальное время задержки между записями в лог
var maxSleepingTime int64 = 5000

// elasticsearch  URL
var elasticURL = "http://localhost:9200"

// elasticsearch  host
var elasticHost = "localhost"

// логгер для вывода на экран
var stdoutLog = logrus.New()

// логгер для вывода в файл
var fileLog = logrus.New()

// флаг, добавлен ли elasticsearch хуг
var elasticHookAdded = false

func main() {
	// считываем переменные окружения, где установлены расположение файла лога,
	// максимальное количество записей, и максимальная задержка между записями в лог.
	readEnvironmentVariables()

	// счетчик записей в логе
	var recordCounter int64 = 0
	// Пока пользователь не нажал Ctrl-C выполняем Вечный цикл.
	for {
		// Время от времени уничтожаем лог
		rotateLog(recordCounter)
		recordCounter++

		// Инициализируем логгер
		initLogger()

		// номер рейса
		flightNumber := rand.Int31n(10)

		// Вычисляем время задержки
		sleepTime := time.Duration(rand.Int63n(maxSleepingTime)) * time.Millisecond

		// Добавляем линию  в лог файл
		addLineToLog(recordCounter, flightNumber, sleepTime)

		// Ждём
		time.Sleep(sleepTime)
	}
}

// Время от времени уничтожаем лог
func rotateLog(recordCounter int64) {
	if recordCounter%maxRecords == 0 {
		err := os.Remove("./logs/" + logFileName)
		if err != nil {
			fmt.Println(err)
		}

	}
}

// считываем переменные окружения, где установлены расположение файла лога,
// максимальное количество записей, и максимальная задержка между записями в лог.
func readEnvironmentVariables() {
	s, ok := os.LookupEnv("MAX_DELAY")
	if ok {
		t, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			fmt.Println("ERROR Can't parse MAX_DELAY as int64")
		} else {
			maxSleepingTime = t
		}
	}
	s, ok = os.LookupEnv("LOG_FILE")
	if ok {
		if s != "" {
			logFileName = s
		}
	}
	s, ok = os.LookupEnv("ELASTIC_URL")
	if ok {
		if s != "" {
			elasticURL = s
		}
	}
	s, ok = os.LookupEnv("ELASTIC_HOST")
	if ok {
		if s != "" {
			elasticHost = s
		}
	}
	s, ok = os.LookupEnv("MAX_RECORDS")
	if ok {
		t, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			fmt.Println("ERROR Can't parse MAX_RECORDS as int64")
		} else {
			maxRecords = t
		}
	}
}

// Инициализируем логгеры
func initLogger() {
	addElasticHookToLogger(stdoutLog)

	timeFormat := "2006-01-02T15:04:05.999Z"
	stdoutLog.SetFormatter(&logrus.JSONFormatter{TimestampFormat: timeFormat})
	fileLog.SetFormatter(&logrus.JSONFormatter{TimestampFormat: timeFormat})

	// Output to stdout instead of the default stderr
	// logrus.SetOutput(os.Stdout)
	stdoutLog.Out = os.Stdout

	// Устанавливаем вывод в файл
	logFile, err := os.OpenFile("./logs/"+logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		// logrus.SetOutput(file)
		fileLog.Out = logFile
	} else {
		logrus.Info("Вывод в файл невозможен, используем stdout")
		fileLog.Out = os.Stdout
	}

	// Only log the warning severity or above.
	// logrus.SetLevel(logrus.WarnLevel)
}

// Добавляет строку в файл лога
func addLineToLog(counter int64, flightNumber int32, delay time.Duration) {
	threshold := time.Duration(maxSleepingTime) * 1000000 * 4 / 5
	msg := fmt.Sprintf("Рейс %d задерживается на %v", flightNumber, delay)
	fields := logrus.Fields{
		"counter":       counter,
		"flight_number": flightNumber,
		"status":        choose(delay > threshold, "DELAYED", "INFO"),
		"wait":          int64(delay / 1000000),
	}
	// если задержка больше порогового значения логируем ошибку,
	// в противном случае логируем сообщение
	if delay > threshold {
		fileLog.WithFields(fields).Error(msg)
		stdoutLog.WithFields(fields).Error(msg)
	} else {
		fileLog.WithFields(fields).Info(msg)
		stdoutLog.WithFields(fields).Info(msg)
	}

}

// addElasticHookToLogger добавляем хуг к логгеру для вывода в Эластик
func addElasticHookToLogger(logger *logrus.Logger) {
	// Если хуг был уже добавлен выходим
	if elasticHookAdded {
		return
	}
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{elasticURL},
	})
	if err != nil {
		fmt.Println("elasticsearch.NewClient error:", err)
		return
	}
	// hook, err := elogrus.NewAsyncElasticHook(client, "localhost", logrus.DebugLevel, "logrus")
	fmt.Println(elasticURL, elasticHost)
	hook, err := elogrus.NewAsyncElasticHook(client, elasticHost, logrus.GetLevel(), "logrus")
	if err != nil {
		fmt.Println("elogrus.NewAsyncElasticHook error 1:", err)
		return
	}
	// possibility to remove hooks
	// https://github.com/sirupsen/logrus/issues/701
	logger.Hooks.Add(hook)
	elasticHookAdded = true
	fmt.Println("Hook to elasticsearch is added.")
}

// Возвращает одну из строк в зависимости от условия
func choose(condition bool, s1, s2 interface{}) interface{} {
	if condition {
		return s1
	}
	return s2
}
