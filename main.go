// Copyright 2020 The RNG Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
  "encoding/gob"
  "flag"
  "fmt"
  "image/color"
  "math"
  "os"
  "time"
  "io/ioutil"

  "gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
  "gonum.org/v1/gonum/integrate"
	"github.com/gonum/stat/distuv"
)

var (
  // Samples number of sample to take
  Samples = flag.Int("samples", 1024*1024, "number of samples to take")
  // Experiment experiment mode
  Experiment = flag.Bool("experiment", false, "experiment mode")
)

// GetSamples gets samples from the rng
func GetSamples() ([200]uint64, []byte, plotter.Values) {
  input, err := os.Open("/dev/TrueRNG")
  if err != nil {
    panic(err)
  }
  defer input.Close()
  buffer := make([]byte, 256)
  n, err := input.Read(buffer)
  histogram, sum, count, samples := [200]uint64{}, 0, 0, make([]byte, 0, *Samples)
  v := make(plotter.Values, 0, *Samples)
Outer:
  for err == nil {
    for _, b := range buffer[:n] {
      for i := 0; i < 8; i++ {
        if b & 1 == 1 {
          sum += 1
        }
        b >>= 1
        count++
        if count == 200 {
          histogram[sum]++
          v = append(v, float64(sum) - 100)
          samples = append(samples, byte(sum))
          if len(samples) == *Samples {
            break Outer
          }
          sum, count = 0, 0
        }
      }
    }
    n, err = input.Read(buffer)
  }

  return histogram, samples, v
}

func y(s, x float64) float64 {
  xx, f := make([]float64, 1000), make([]float64, 1000)
  step, t := x / 1000, 0.0
  for i := 0; i < 1000; i++ {
    xx[i], f[i] = t, math.Pow(t, s-1) * math.Exp(-t)
    t += step
  }
  return integrate.Trapezoidal(xx, f)
}

func main() {
  flag.Parse()

  if *Experiment {
    start := time.Now()
    reference := [200]uint64{}
    input, err := os.Open("histogram.bin")
    if err != nil {
      panic(err)
    }
    defer input.Close()
    decoder := gob.NewDecoder(input)
    err = decoder.Decode(&reference)
    if err != nil {
      panic(err)
    }
    histogram, _, _ := GetSamples()
    x2 := 0.0
    for i, x := range histogram {
      m := float64(reference[i])
      d := float64(x) - m
      if m != 0 {
        x2 += d*d/m
      }
    }
    fmt.Println("X^2", x2)
    k := float64(200)
    pvalue := 1 - y(k/2, x2/2)/math.Gamma(k/2)
    fmt.Println("P-value", pvalue)
    fmt.Println(time.Now().Sub(start))
    return
  }

  start := time.Now()
  histogram, samples, v := GetSamples()
  fmt.Println(histogram)
  fmt.Println(time.Now().Sub(start))

  output, err := os.Create("histogram.bin")
  if err != nil {
    panic(err)
  }
  defer output.Close()
  encoder := gob.NewEncoder(output)
  err = encoder.Encode(histogram)
  if err != nil {
    panic(err)
  }

  err = ioutil.WriteFile("samples.bin", samples, 0644)
  if err != nil {
    panic(err)
  }

  s, ss, length := 0.0, 0.0, float64(len(v))
  for _, value := range v {
    s += value
    ss += value * value
  }
  std := math.Sqrt((ss - s*s/length) / length)
  for i, value := range v {
    v[i] = value / std
  }

	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Histogram"

	h, err := plotter.NewHist(v, 200)
	if err != nil {
		panic(err)
	}
	h.Normalize(.3)
	p.Add(h)

	norm := plotter.NewFunction(distuv.UnitNormal.Prob)
	norm.Color = color.RGBA{R: 255, A: 255}
	norm.Width = vg.Points(2)
	p.Add(norm)

  err = p.Save(8*vg.Inch, 8*vg.Inch, "histogram.png")
	if err != nil {
		panic(err)
	}
}
