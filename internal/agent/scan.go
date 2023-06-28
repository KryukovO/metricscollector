package agent

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/KryukovO/metricscollector/internal/metric"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type ScanResult struct {
	mtrc *metric.Metrics
	err  error
}

func scanMetrics(ctx context.Context) chan ScanResult {
	channels := []chan ScanResult{
		scanRuntimeMetrics(ctx, rand.New(rand.NewSource(time.Now().UnixNano()))),
		scanPSUtilMetrics(ctx),
	}

	outCh := make(chan ScanResult, len(channels))
	wg := new(sync.WaitGroup)

	for _, ch := range channels {
		wg.Add(1)

		go func(inCh <-chan ScanResult) {
			defer wg.Done()

			for data := range inCh {
				select {
				case <-ctx.Done():
					return
				case outCh <- data:
				}
			}
		}(ch)
	}

	go func() {
		wg.Wait()

		close(outCh)
	}()

	return outCh
}

func scanRuntimeMetrics(ctx context.Context, rnd *rand.Rand) chan ScanResult {
	outCh := make(chan ScanResult, 1)

	go func() {
		defer close(outCh)

		buf := make(map[string]interface{})

		rndVal := float64(rand.Int())
		if rnd != nil {
			rndVal = float64(rnd.Int())
		}

		buf["RandomValue"] = rndVal

		rtm := &runtime.MemStats{}

		runtime.ReadMemStats(rtm)

		buf["Alloc"] = float64(rtm.Alloc)
		buf["BuckHashSys"] = float64(rtm.BuckHashSys)
		buf["Frees"] = float64(rtm.Frees)
		buf["GCCPUFraction"] = rtm.GCCPUFraction
		buf["GCSys"] = float64(rtm.GCSys)
		buf["HeapAlloc"] = float64(rtm.HeapAlloc)
		buf["HeapIdle"] = float64(rtm.HeapIdle)
		buf["HeapInuse"] = float64(rtm.HeapInuse)
		buf["HeapObjects"] = float64(rtm.HeapObjects)
		buf["HeapReleased"] = float64(rtm.HeapReleased)
		buf["HeapSys"] = float64(rtm.HeapSys)
		buf["LastGC"] = float64(rtm.LastGC)
		buf["Lookups"] = float64(rtm.Lookups)
		buf["MCacheInuse"] = float64(rtm.MCacheInuse)
		buf["MCacheSys"] = float64(rtm.MCacheSys)
		buf["MSpanInuse"] = float64(rtm.MSpanInuse)
		buf["MSpanSys"] = float64(rtm.MSpanSys)
		buf["Mallocs"] = float64(rtm.Mallocs)
		buf["NextGC"] = float64(rtm.NextGC)
		buf["NumForcedGC"] = float64(rtm.NumForcedGC)
		buf["NumGC"] = float64(rtm.NumGC)
		buf["OtherSys"] = float64(rtm.OtherSys)
		buf["PauseTotalNs"] = float64(rtm.PauseTotalNs)
		buf["StackInuse"] = float64(rtm.StackInuse)
		buf["StackSys"] = float64(rtm.StackSys)
		buf["Sys"] = float64(rtm.Sys)
		buf["TotalAlloc"] = float64(rtm.TotalAlloc)

		for mName, mVal := range buf {
			mtrc, err := metric.NewMetrics(mName, "", mVal)
			data := ScanResult{mtrc: mtrc, err: err}

			select {
			case <-ctx.Done():
				return
			case outCh <- data:
			}
		}
	}()

	return outCh
}

func scanPSUtilMetrics(ctx context.Context) chan ScanResult {
	outCh := make(chan ScanResult, 1)

	go func() {
		defer close(outCh)

		buf := make(map[string]interface{})

		vmStat, err := mem.VirtualMemory()
		if err != nil {
			outCh <- ScanResult{err: err}

			return
		}

		buf["TotalMemory"] = float64(vmStat.Total)
		buf["FreeMemory"] = float64(vmStat.Free)

		cpuStat, err := cpu.Times(true)
		if err != nil {
			outCh <- ScanResult{err: err}

			return
		}

		for i, ts := range cpuStat {
			buf[fmt.Sprintf("CPUutilization%d", i)] = ts.Idle
		}

		for mName, mVal := range buf {
			mtrc, err := metric.NewMetrics(mName, "", mVal)
			data := ScanResult{mtrc: mtrc, err: err}

			select {
			case <-ctx.Done():
				return
			case outCh <- data:
			}
		}
	}()

	return outCh
}
