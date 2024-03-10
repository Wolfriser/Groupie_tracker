package pkg_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"lzhuk/groupie-tracker/pkg"
)

func TestMain(m *testing.M) {
	// Установка текущего рабочего каталога в корневой каталог проекта
	err := os.Chdir("../")
	if err != nil {
		log.Fatal(err)
	}

	// Запуск тестов
	code := m.Run()

	// Очистка и завершение тестов
	os.Exit(code)
}

// Тест 1 для проверки функции обработчика домашней страницы
func TestHomeHandler(t *testing.T) {

	// Подтест 1.1 для проверки работы функции обработчика с правильными данными запроса
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(pkg.HomeHandler)

	handler.ServeHTTP(rr, req)

	// Проверка кода состояния 200
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Ожидался статус %v, но получен %v", http.StatusOK, status)
	}

	// Подтест 1.2 для проверки работы функции обработчика с неправильными URL в запросе 
	req, err = http.NewRequest("GET", "/k", nil) // Вместо URL = / передан /k
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(pkg.HomeHandler)

	handler.ServeHTTP(rr, req)

	// Проверка кода состояния 404
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Ожидался статус %v, но получен %v", http.StatusNotFound, status)
	}

	// Подтест 1.3 для проверки работы функции обработчика с неправильными методом в запросе 
	req, err = http.NewRequest("POST", "/", nil) // Вместо метода GET передан метод POST
	if err != nil {
		t.Fatal(err)
	}

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(pkg.HomeHandler)

	handler.ServeHTTP(rr, req)

	// Проверка кода состояния 405
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Ожидался статус %v, но получен %v", http.StatusMethodNotAllowed, status)
	}
}


// Тест 2 для проверки функции обработчика страницы с кодами ошибок
func TestErrorHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(pkg.HomeHandler)

	handler.ServeHTTP(rr, req)

	// Проверка кода состояния
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Ожидался статус %v, но получен %v", http.StatusMethodNotAllowed, status)
	}
}
