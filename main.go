package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"lzhuk/groupie-tracker/pkg"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	isDataArtistWritten    bool // Флаг, указывающий, была ли уже запись данных с артистами и группами
	isDataRelationsWritten bool // Флаг, указывающий, была ли уже запись данных со связями
	isDataLocationsWritten bool
	Mux                    *http.ServeMux
)

func checkInternetConnection() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://clients3.google.com/generate_204", nil)
	if err != nil {
		return false
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	isDataArtistWritten = false
	isDataRelationsWritten = false
	isDataLocationsWritten = false

	return resp.StatusCode == http.StatusNoContent
}

func main() {
	// Инициализация журнала логирования
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		log.Println("Не удалось открыть файл лога:", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetPrefix("Сформирована запись: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Создаем родительский контекст
	parentCtx := context.Background()

	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel() // Отложенный вызов cancel для освобождения ресурсов

	// Проверяем наличие файла кеша с артистами и группами при запуске сервера
	if _, err := os.Stat("cacheArtist.json"); err == nil {
		// Если файл существует, загружаем данные из него в кеш
		data, err := ioutil.ReadFile("cacheArtist.json")
		if err != nil {
			log.Println("Ошибка при чтении файла кэша с артистами и группами:", err)
		} else {
			err = json.Unmarshal(data, &pkg.ResponseData.Band)
			if err != nil {
				log.Println("Ошибка при декодировании данных  с артистами и группами из файла кэша:", err)
			} else {
				log.Println("Данные с артистами и группами из файла кэша успешно загружены")
			}
		}
	} else {
		log.Println("Файл с кэшем артистов и групп отсутствует")
	}

	// Проверяем наличие файла кеша с связями при запуске сервера
	if _, err := os.Stat("cacheRelation.json"); err == nil {
		// Если файл существует, загружаем данные из него в кеш
		data, err := ioutil.ReadFile("cacheRelation.json")
		if err != nil {
			log.Println("Ошибка при чтении файла кэша со связями:", err)
		} else {
			err = json.Unmarshal(data, &pkg.RelationInfo)
			if err != nil {
				log.Println("Ошибка при декодировании данных о связях из файла кэша:", err)
			} else {
				log.Println("Данные со связями из файла кэша успешно загружены")
			}
		}
	} else {
		log.Println("Файл со связями отсутствует")
	}

	// Проверяем наличие файла кеша с локациями при запуске сервера
	if _, err := os.Stat("cacheLocation.json"); err == nil {
		// Если файл существует, загружаем данные из него в кеш
		data, err := ioutil.ReadFile("cacheLocation.json")
		if err != nil {
			log.Println("Ошибка при чтении файла кэша с локациями:", err)
		} else {
			err = json.Unmarshal(data, &pkg.RelationInfo)
			if err != nil {
				log.Println("Ошибка при декодировании данных о локациях из файла кэша:", err)
			} else {
				log.Println("Данные о локациях из файла кэша успешно загружены")
			}
		}
	} else {
		log.Println("Файл с локациями отсутствует")
	}

	// Запускаем отдельную горутину для проверки соединения с интернетом
	go timerInternetConnect(ctx)

	// Загрузка данных из API в базу данных и кэш
	pkg.UpdateCache()

	// Обновление данных в кэше
	go timerCache(ctx)

	// Запускаем сервер
	server := Server()
	go serverStart(ctx, server)

	// Ждем сигнала остановки сервера
	<-stop

	fmt.Println("\n" + "Завершение текущих процессов...")

	time.Sleep(3 * time.Second)

	// Отменяем контекст, чтобы завершить работу сервера
	cancel()

	time.Sleep(1 * time.Second)
}

func timerInternetConnect(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	x := 1
	counter := 180

	for x != 0 {
		select {
		case <-ticker.C:
			if checkInternetConnection() && !isDataArtistWritten && !isDataRelationsWritten {
				log.Println("Соединение с интернетом есть")
			} else {
				log.Println("Соединение с интернетом отсутствует")

				if counter%180 == 0 {
					cacheArtistJSON, err := json.Marshal(pkg.ResponseData.Band)
					if err != nil {
						log.Println("Ошибка при преобразовании данных об артистах и группах в JSON:", err)
						return
					}
					if err := pkg.SaveCacheToFile("cacheArtist.json", cacheArtistJSON); err != nil {
						log.Println("Ошибка при сохранении данных об артистах и группах в файл:", err)
					} else {
						log.Println("Данные об артистах и группах успешно сохранены в файл")
						isDataArtistWritten = true
					}

					cacheRelationJSON, err := json.Marshal(pkg.RelationInfo)
					if err != nil {
						log.Println("Ошибка при преобразовании данных о связях в JSON:", err)
						return
					}
					if err := pkg.SaveCacheToFile("cacheRelation.json", cacheRelationJSON); err != nil {
						log.Println("Ошибка при сохранении данных о связях в файл:", err)
					} else {
						log.Println("Данные о связях успешно сохранены в файл")
						isDataRelationsWritten = true
					}

					cacheLocationJSON, err := json.Marshal(pkg.LocationInfo)
					if err != nil {
						log.Println("Ошибка при преобразовании данных о локации в JSON:", err)
						return
					}
					if err := pkg.SaveCacheToFile("cacheLocation.json", cacheLocationJSON); err != nil {
						log.Println("Ошибка при сохранении данных о локации в файл:", err)
					} else {
						log.Println("Данные о локациях успешно сохранены в файл")
						isDataLocationsWritten = true
					}
				}
				counter += 1
			}
		case <-ctx.Done():
			x = 0
			cacheArtistJSON, err := json.Marshal(pkg.ResponseData.Band)
			if err != nil {
				log.Println("Ошибка при преобразовании данных об артистах и группах в JSON:", err)
				return
			}
			if err := pkg.SaveCacheToFile("cacheArtist.json", cacheArtistJSON); err != nil {
				log.Println("Ошибка при сохранении данных об артистах и группах в файл:", err)
			} else {
				log.Println("Данные об артистах и группах успешно сохранены в файл")
				isDataArtistWritten = true
			}
			cacheRelationJSON, err := json.Marshal(pkg.RelationInfo)
			if err != nil {
				log.Println("Ошибка при преобразовании данных о связях в JSON:", err)
				return
			}
			if err := pkg.SaveCacheToFile("cacheRelation.json", cacheRelationJSON); err != nil {
				log.Println("Ошибка при сохранении данных о связях в файл:", err)
			} else {
				log.Println("Данные о связях успешно сохранены в файл")
				isDataRelationsWritten = true
			}
			cacheLocationJSON, err := json.Marshal(pkg.LocationInfo)
			if err != nil {
				log.Println("Ошибка при преобразовании данных о локации в JSON:", err)
				return
			}
			if err := pkg.SaveCacheToFile("cacheLocation.json", cacheLocationJSON); err != nil {
				log.Println("Ошибка при сохранении данных о локации в файл:", err)
			} else {
				log.Println("Данные о локациях успешно сохранены в файл")
				isDataLocationsWritten = true
			}
		}
	}
}

func serverStart(ctx context.Context, server *http.Server) {
	go func() {
		<-ctx.Done()

		fmt.Println("Завершение текущих запросов...")

		// Создаем контекст с тайм-аутом для завершения оставшихся запросов
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Останавливаем сервер и ждем завершения всех запросов
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Fatal("Ошибка при завершении работы сервера:", err)
		} else {
			log.Println("Сервер успешно остановлен")
			fmt.Println("Сервер успешно остановлен")
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Ошибка при запуске сервера:", err)
	}
}

func timerCache(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(60 * time.Second):
				pkg.UpdateCache()
			}
		}
	}()
}

func Server() *http.Server {
	Mux = http.NewServeMux()

	Mux.HandleFunc("/", pkg.HomeHandler)

	Mux.HandleFunc("/band", pkg.BandHandler)

	Mux.HandleFunc("/search", pkg.SearchHandler)

	fileServer := http.FileServer(http.Dir("web/static"))

	Mux.Handle("/web/static/", http.StripPrefix("/web/static/", fileServer))

	S := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      Mux,
	}
	log.Println("Сервер успешно запущен")
	fmt.Printf("Cервер успешно запущен: %s"+"\n", "http://localhost:8080")

	return S
}
