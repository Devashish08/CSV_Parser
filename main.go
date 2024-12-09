package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// problem represents a quiz question and its answer.
type problem struct {
	question string
	answer   string
}

func main() {
	csvFileName := flag.String("csv", "problems.csv", "a csv file in the format of 'question,answer'")
	timeLimit := flag.Int("limit", 30, "the time limit for the quiz in seconds")
	flag.Parse()

	problems, err := loadProblems(*csvFileName)
	if err != nil {
		exit(err.Error())
	}

	score := runQuiz(problems, *timeLimit, os.Stdin)
	fmt.Printf("You scored %d out of %d.\n", score, len(problems))
}

// loadProblems reads and parses the CSV file into a slice of problems.
func loadProblems(fileName string) ([]problem, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open the CSV file: %s", fileName)
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.FieldsPerRecord = -1 // Allow variable number of fields per record

	var problems []problem
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse the provided CSV file: %v", err)
		}
		if len(record) < 2 {
			continue // skip invalid records
		}
		problems = append(problems, problem{
			question: strings.TrimSpace(record[0]),
			answer:   strings.TrimSpace(record[1]),
		})
	}
	return problems, nil
}

// runQuiz conducts the quiz, returns the number of correct answers.
func runQuiz(problems []problem, timeLimit int, input io.Reader) int {
	timer := time.NewTimer(time.Duration(timeLimit) * time.Second)
	defer timer.Stop()

	correct := 0
	scanner := bufio.NewScanner(input)

problemloop:
	for i, p := range problems {
		fmt.Printf("Problem #%d: %s = ", i+1, p.question)

		answerCh := make(chan string)

		go func() {
			if scanner.Scan() {
				answerCh <- strings.TrimSpace(scanner.Text())
			} else {
				close(answerCh)
			}
		}()

		select {
		case <-timer.C:
			fmt.Println("\nTime's up!")
			break problemloop
		case answer, ok := <-answerCh:
			if !ok {
				break problemloop
			}
			if strings.EqualFold(answer, p.answer) {
				correct++
			}
		}
	}

	return correct
}

// exit prints the error message and exits the program.
func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
