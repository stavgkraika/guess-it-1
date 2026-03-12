# guess-it-1

A streaming prediction interval generator that dynamically adapts to time-series data patterns.

## Overview

This program reads a stream of integer values and produces prediction intervals for upcoming values. It uses adaptive algorithms that learn from prediction accuracy to optimize interval width.

## Features

- **Adaptive Intervals**: Dynamically adjusts prediction width based on hit rate feedback
- **Robust Statistics**: Uses both standard deviation and MAD for outlier resistance
- **Trend Tracking**: Combines momentum and mean reversion for forecasting
- **Efficient Processing**: Ring buffers and buffered I/O for high-performance streaming

## Algorithm

The predictor maintains:
- Rolling window of recent values (size 50)
- Absolute differences between consecutive values
- Hit/miss history for recent predictions (size 20)

For each new value, it:
1. Evaluates previous prediction accuracy
2. Adjusts interval width multiplier (k) based on hit rate
3. Computes statistical measures (mean, std dev, MAD)
4. Forecasts next value using trend + mean reversion
5. Generates interval bounds with adaptive width

## Building

```bash
go build -o guess-it-1 ./cmd
```

## Usage

### Run without building

```bash
go run ./cmd < input.txt > output.txt
```

### Run with build

```bash
./guess-it-1 < input.txt > output.txt
```

Input: Stream of integers (one per line)  
Output: Prediction intervals as `lower upper` pairs (one per line)

## Requirements

- Go 1.21 or higher
