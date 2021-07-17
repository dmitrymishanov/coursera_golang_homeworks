package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

var Md5Mutex = sync.Mutex{}

func LockingDataSignerMd5(value string) string {
	Md5Mutex.Lock()
	result := DataSignerMd5(value)
	Md5Mutex.Unlock()
	return result
}

func ChanneledDataSignerCrc32(value string) chan string {
	ch := make(chan string)
	go func(ch chan string) {
		ch <- DataSignerCrc32(value)
	}(ch)
	return ch
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for value := range in {
		wg.Add(1)
		go func(value interface{}, out chan interface{}) {
			defer wg.Done()
			strValue := fmt.Sprint(value)

			md5Value := LockingDataSignerMd5(strValue)

			var left, right string
			leftCh := ChanneledDataSignerCrc32(strValue)
			rightCh := ChanneledDataSignerCrc32(md5Value)

			for i := 0; i < 2; i++ {
				select {
				case left = <-leftCh:
				case right = <-rightCh:
				}
			}
			out <- left + "~" + right

		}(value, out)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for value := range in {
		wg.Add(1)
		go func(value interface{}, out chan interface{}) {
			defer wg.Done()
			strValue := value.(string)
			parts := make([]string, 6)
			innerWg := &sync.WaitGroup{}
			for i := 0; i <= 5; i++ {
				innerWg.Add(1)
				go func(i int) {
					defer innerWg.Done()
					part := DataSignerCrc32(fmt.Sprint(i, strValue))
					parts[i] = part
				}(i)
			}
			innerWg.Wait()
			out <- strings.Join(parts, "")
		}(value, out)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	var result []string
	for value := range in {
		result = append(result, value.(string))
	}

	sort.Strings(result)
	out <- strings.Join(result, "_")
}

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	in := make(chan interface{}, 100)
	defer close(in)
	for _, jobItem := range jobs {
		out := make(chan interface{}, 100)
		wg.Add(1)
		go func(jobItem job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			jobItem(in, out)
		}(jobItem, in, out)
		in = out
	}
}
