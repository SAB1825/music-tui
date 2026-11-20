package visualizer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
	"strconv"
)

const fullScale = 1000

func parseFrame(line []byte, bars int) ([]float64, error) {
	line = bytes.TrimSuffix(line, []byte(";"))
	frame := make([]float64, 0, bars)
	for field := range bytes.SplitSeq(line, []byte(";")) {
		value, err := strconv.Atoi(string(field))
		if err != nil || value < 0 || value > fullScale || len(frame) == bars {
			return nil, fmt.Errorf("invalid CAVA spectrum sample")
		}
		frame = append(frame, float64(value)/fullScale)
	}
	if len(frame) != bars {
		return nil, fmt.Errorf("CAVA returned %d bars, expected %d", len(frame), bars)
	}
	return frame, nil
}

func readFrames(r io.Reader, bars int, emit func([]float64)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024), 4096)
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if n := bytes.IndexByte(data, '\n'); n >= 0 {
			return n + 1, data[:n], nil
		}
		if atEOF {
			return len(data), nil, nil
		}
		return 0, nil, nil
	})
	var peak []float64
	for scanner.Scan() {
		frame, err := parseFrame(scanner.Bytes(), bars)
		if err != nil {
			return err
		}
		for i, level := range peak {
			frame[i] = max(frame[i], level-.075)
		}
		if slices.Equal(peak, frame) {
			continue
		}
		peak = append(peak[:0], frame...)
		emit(frame)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read CAVA frame: %w", err)
	}
	return nil
}
