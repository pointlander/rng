// Copyright 2020 The RNG Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
  "fmt"
  "image/color"
  "math"
  "os"

  "gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"github.com/gonum/stat/distuv"
)

func main() {
  input, err := os.Open("/dev/TrueRNG")
  if err != nil {
    panic(err)
  }
  buffer := make([]byte, 256)
  n, err := input.Read(buffer)
  histogram, sum, count, samples := [200]uint64{}, 0, 0, 0
  v := make(plotter.Values, 0, 256)
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
          samples++
          if samples == 30000 {
            break Outer
          }
          sum, count = 0, 0
        }
      }
    }
    n, err = input.Read(buffer)
  }
  fmt.Println(histogram)

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
