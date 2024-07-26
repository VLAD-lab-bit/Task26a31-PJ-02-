package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	BufferSize = 5
	Interval   = 5 * time.Second
)

func filterNegative(done <-chan struct{}, input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for {
			select {
			case <-done:
				return
			case i, isChannelOpen := <-input:
				if !isChannelOpen {
					return
				}
				if i >= 0 {
					output <- i
				}
			}
		}
	}()
	return output
}

func filterNonMultipleOfThree(done <-chan struct{}, input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for {
			select {
			case <-done:
				return
			case i, isChannelOpen := <-input:
				if !isChannelOpen {
					return
				}
				if i != 0 && i%3 == 0 {
					output <- i
				}
			}
		}
	}()
	return output
}

type RingBuffer struct {
	data    []int
	maxSize int
	nextIn  int
	nextOut int
	count   int
	mu      sync.Mutex
}

func NewRingBuffer(maxSize int) *RingBuffer {
	return &RingBuffer{
		data:    make([]int, maxSize),
		maxSize: maxSize,
		nextIn:  0,
		nextOut: 0,
		count:   0,
	}
}

func (rb *RingBuffer) Push(val int) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if rb.count == rb.maxSize {
		return false
	}
	rb.data[rb.nextIn] = val
	rb.nextIn = (rb.nextIn + 1) % rb.maxSize
	rb.count++
	return true
}

func (rb *RingBuffer) Pop() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	val := rb.data[rb.nextOut]
	rb.nextOut = (rb.nextOut + 1) % rb.maxSize
	rb.count--
	return val
}

func (rb *RingBuffer) Count() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}

func dataSource(done chan<- struct{}) <-chan int {
	output := make(chan int)
	scanner := bufio.NewScanner(os.Stdin)
	go func() {
		defer close(output)
		for scanner.Scan() {
			input := scanner.Text()
			if input == "exit" {
				close(done)
				return
			}
			num, err := strconv.Atoi(input)
			if err == nil {
				output <- num
			} else {
				fmt.Println("Введено нечисловое значение, игнорируется:", input)
			}
		}
	}()
	return output
}

func dataConsumer(done <-chan struct{}, input <-chan int, bufferSize int, interval time.Duration) {
	buffer := NewRingBuffer(bufferSize)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case val, isOpen := <-input:
			if !isOpen {
				return
			}
			if !buffer.Push(val) {
				fmt.Println("Буфер переполнен, значение игнорируется:", val)
			}
		case <-ticker.C:
			for buffer.Count() > 0 {
				fmt.Println("Получены данные:", buffer.Pop())
			}
		}
	}
}

func main() {
	done := make(chan struct{})
	pipeline := filterNonMultipleOfThree(done, filterNegative(done, dataSource(done)))
	go dataConsumer(done, pipeline, BufferSize, Interval)
	<-done
	fmt.Println("Программа завершила работу.")
}
