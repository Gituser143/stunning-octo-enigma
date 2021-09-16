package load

import (
	"log"
	"math/rand"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

func (sc *StressClient) getDistribution(distributionType string, steps int, minRate int, maxRate int) []int {

	result := make([]int, steps)

	s1 := rand.NewSource(time.Hour.Nanoseconds())
	r1 := rand.New(s1)

	switch distributionType {
	case "inc":
		step := (maxRate - minRate) / (steps - 1)
		for i := 0; i < steps; i++ {
			result[i] = minRate + step*i
		}
	case "dec":
		step := (maxRate - minRate) / (steps - 1)
		for i := 0; i < steps; i++ {
			result[i] = maxRate - step*i
		}
	case "zipf":
		zipf := rand.NewZipf(r1, 1.1, 100, uint64(maxRate))
		for i := 0; i < steps; i++ {
			result[i] = minRate + int(zipf.Uint64())
		}
	case "uni":
		for i := 0; i < steps; i++ {
			result[i] = minRate + r1.Intn(maxRate-minRate)
		}
	}

	return result
}

// StressApplication stress tests the application using a given number of
// workers, for a specified duration, with minimum and maximum rate of requests
// sent for iterations specified by steps.
func (sc *StressClient) StressApplication(
	distributionType string,
	steps int,
	duration int,
	workers int,
	minRate int,
	maxRate int,
) {
	distribution := sc.getDistribution(distributionType, steps, minRate, maxRate)
	targets := sc.getTargets()

	for _, frequency := range distribution {
		rate := vegeta.Rate{Freq: frequency, Per: time.Second}
		attackerFunc := vegeta.Workers(uint64(workers))
		attacker := vegeta.NewAttacker(attackerFunc)
		targeter := vegeta.NewStaticTargeter(targets...)
		res := attacker.Attack(targeter, rate, time.Duration(duration)*time.Second, "")
		open := true
		var result *vegeta.Result

		for open {
			result, open = <-res
			if result != nil {
				log.Println("response received", result.Code)
			} else {
				log.Println("nil result")
			}
		}
	}
}