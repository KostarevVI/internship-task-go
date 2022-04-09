package main

import (
	"fmt"
	"sync"
)

type sensorData struct {
	value int
	isNew bool
}

type SensorDataCollection struct {
	Size        int            `json:"size,omitempty"`
	Data        []sensorData   `json:"data,omitempty"`
	Average     float32        `json:"average"`
	IsCorrupted bool           `json:"isCorrupted"`
	Wg          sync.WaitGroup `json:"wg,omitempty"`
}

// InitCollection fills fields
func (c *SensorDataCollection) InitCollection(size int) {
	c.Size = size
	c.Data = make([]sensorData, size)
	c.Average = -1
	c.IsCorrupted = true
	c.Wg = sync.WaitGroup{}

	for i := range c.Data {
		c.Data[i].value = -1
	}
}

// ResetFlags is called when we're receiving new values and old results became not relevant
func (c *SensorDataCollection) ResetFlags() {
	for i := 0; i < c.Size; i++ {
		c.Data[i].isNew = false
	}
	c.IsCorrupted = true

	fmt.Println("Reset | Avr:", c.Average, ", IsCorrupted:", c.IsCorrupted)
}

// CalculateAverage is calculating average (wow) with IsCorrupted flag
func (c *SensorDataCollection) CalculateAverage() {
	go func() {
		c.Wg.Wait()
		sum := 0
		amount := 0
		cFlag := false
		for _, sensorData := range c.Data {
			if sensorData.value >= 0 {
				sum += sensorData.value
				if sensorData.isNew == false {
					cFlag = true
				}
				amount++
			}
		}
		if amount > 0 {
			c.Average = float32(sum) / float32(amount)
			c.IsCorrupted = cFlag

			fmt.Println("Calc | Avr:", c.Average, ", IsCorrupted:", c.IsCorrupted)
		}

		return
	}()
}

// UpdateData uploads new values from sensors and rises flags isNew
func (c *SensorDataCollection) UpdateData(index int, value int, isNew bool) {
	c.Data[index].value = value
	c.Data[index].isNew = isNew
}
