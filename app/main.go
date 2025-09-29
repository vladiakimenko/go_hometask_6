package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type taskType int

const (
	IO taskType = iota
	CPU
)

func (t taskType) String() string {
	switch t {
	case IO:
		return "IO"
	case CPU:
		return "CPU"
	default:
		panic(fmt.Sprintf("unknown task type: %d", t))
	}
}

var idIncrementor int = 0

type Task struct {
	ID   int
	Name string
	Type taskType
}

type Result struct {
	TaskID   int
	TaskName string
	Success  bool
	Duration time.Duration
	Message  string
}

// simulateIOWork имитирует I/O операцию (чтение файла, сетевой запрос)
func simulateIOWork(task Task, wg *sync.WaitGroup, results chan<- Result) {
	fmt.Printf("Начинаю %s задачу %d: %s\n", task.Type, task.ID, task.Name)
	defer wg.Done()
	duration := time.Duration(rand.Intn(16)+5) * time.Second
	time.Sleep(duration)
	results <- Result{
		TaskID:   task.ID,
		TaskName: task.Name,
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Successfully completed IO task %s", task.Name),
	}
	fmt.Printf("Закончил I/O задачу %d: %s\n", task.ID, task.Name)
}

// simulateComputeWork имитирует вычислительную задачу
func simulateComputeWork(task Task, wg *sync.WaitGroup, results chan<- Result) {
	fmt.Printf("Начинаю %s задачу %d: %s\n", task.Type, task.ID, task.Name)
	defer wg.Done()
	timeStart := time.Now()
	fibonacci(45 + rand.Intn(4))
	duration := time.Since(timeStart)
	results <- Result{
		TaskID:   task.ID,
		TaskName: task.Name,
		Success:  true,
		Duration: duration,
		Message:  fmt.Sprintf("Successfully completed CPU task %s", task.Name),
	}
	fmt.Printf("Закончил вычислительную задачу %d: %s\n", task.ID, task.Name)
}

// fibonacci вычисляет число Фибоначчи (готовая функция)
func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

func monitorProgress(totalTasks int, results <-chan Result, done chan<- bool) {
	completed := 0
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case _, ok := <-results:
			if !ok { // Канал закрыт - все задачи завершены
				fmt.Println("Все задачи завершены")
				done <- true
				return
			}
			completed++
			// для простоты считаем, что нам пофиг на задержку сигнала о завершении до следующего тика
		case <-ticker.C:
			if completed == totalTasks {
				done <- true
				return
			} else {
				fmt.Printf("Завершено %.0f%% задач\n", (float64(completed)/float64(totalTasks))*100)
			}
		}
	}
}

func nextID() int {
	idIncrementor++
	return idIncrementor
}

func main() {
	fmt.Println(" === ДЕМОНСТРАЦИЯ ГОРУТИН В GO ===")
	fmt.Println("Запуск параллельных задач с использованием горутин, WaitGroup и каналов")

	// Инициализация генератора случайных чисел - НЕ ТРЕБУЕТСЯ В GO 1.20+
	// rand.Seed(time.Now().UnixNano())

	ioTasks := []Task{
		{nextID(), "Загрузка данных с удалённого АПИ", IO},
		{nextID(), "Чтение файла", IO},
		{nextID(), "Выборка данных из БД", IO},
		{nextID(), "Сохранение файла в хранилище статики", IO},
	}

	computeTasks := []Task{
		{nextID(), "Вычисление", CPU},
		{nextID(), "Анализ данных", CPU},
		{nextID(), "Преобразование формата фото", CPU},
		{nextID(), "Векторное представление изображения", CPU},
	}

	allTasks := append(ioTasks, computeTasks...)
	totalTasks := len(allTasks)

	results := make(chan Result)
	monitorDone := make(chan bool)

	var wg sync.WaitGroup

	fmt.Printf("Запускаю %d задач:\n", totalTasks)
	for _, task := range allTasks {
		fmt.Printf("   • %s (ID: %d, Тип: %s)\n", task.Name, task.ID, task.Type)
	}

	go monitorProgress(totalTasks, results, monitorDone)

	fmt.Println("\nЗапуск горутин...")
	startTime := time.Now()

	for _, task := range ioTasks {
		wg.Add(1)
		go simulateIOWork(task, &wg, results)
	}

	for _, task := range computeTasks {
		wg.Add(1)
		go simulateComputeWork(task, &wg, results)
	}

	fmt.Printf("Запущено %d горутин\n", len(allTasks))

	wg.Wait()
	close(results)

	<-monitorDone
	totalExecutionTime := time.Since(startTime)

	fmt.Printf("\n === ПРОГРАММА ЗАВЕРШЕНА ===\n")
	fmt.Printf("⏱️  Общее время выполнения программы: %v\n", totalExecutionTime)
	fmt.Printf("Все горутины успешно завершены!\n")
}
