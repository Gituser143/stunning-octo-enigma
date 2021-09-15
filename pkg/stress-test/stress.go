package stressTest

import (
	"fmt"
	"math/rand"
	"time"

	vegeta "github.com/tsenart/vegeta/lib"
)

const host = "localhost"
const protocol = "http"
const webuiPort = "30080"

const (
	seed = 42
)

func checkStatus() vegeta.Target {
	url := fmt.Sprintf("%s://%s:%s/tools.descartes.teastore.webui/status", host, protocol, webuiPort)
	target := vegeta.Target{
		URL:    url,
		Method: "GET",
	}
	return target
}

func checkLogin(username string, password string) vegeta.Target {
	url := fmt.Sprintf("%s://%s:%s/tools.descartes.teastore.webui/loginAction?username=%s&password=%s", host, protocol, webuiPort, username, password)
	target := vegeta.Target{
		URL:    url,
		Method: "GET",
	}
	return target
}

func checkSuccessfulLogin() vegeta.Target {
	return checkLogin("user2", "password")
}

func checkFailedLogin() vegeta.Target {
	return checkLogin("testuser", "password")
}

func getCategoryEndpoints() []vegeta.Target {
	var targets []vegeta.Target
	categoryIds := []int{2, 3, 4, 5, 6}
	for _, categoryId := range categoryIds {
		pageNumbers := []int{1, 2, 3, 4, 5}
		for _, pageNumber := range pageNumbers {
			url := fmt.Sprintf("%s://%s:%s/tools.descartes.teastore.webui/category?category=%d&page=%d", host, protocol, webuiPort, categoryId, pageNumber)
			target := vegeta.Target{
				URL:    url,
				Method: "GET",
			}
			targets = append(targets, target)
		}
	}
	return targets
}

func getProductEndpoints() []vegeta.Target {
	var targets []vegeta.Target
	for productId := 100; productId < 150; productId++ {
		url1 := fmt.Sprintf("%s://%s:%s/tools.descartes.teastore.webui/cartAction?addToCart=&productid=%d", host, protocol, webuiPort, productId)
		url2 := fmt.Sprintf("%s://%s:%s/tools.descartes.teastore.webui/product?id=%d", host, protocol, webuiPort, productId)
		target1 := vegeta.Target{
			URL:    url1,
			Method: "POST",
		}
		target2 := vegeta.Target{
			URL:    url2,
			Method: "GET",
		}
		targets = append(targets, target1)
		targets = append(targets, target2)
	}
	return targets
}

func getTargets() []vegeta.Target {
	var targets []vegeta.Target

	targets = append(targets, checkStatus())
	targets = append(targets, checkFailedLogin())
	targets = append(targets, checkSuccessfulLogin())

	targets = append(targets, getCategoryEndpoints()...)
	targets = append(targets, getProductEndpoints()...)

	return targets
}

func getDistribution(distributionType string, steps int, minRate int, maxRate int) []int {

	result := make([]int, steps)

	s1 := rand.NewSource(seed)
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

func StressTestTeaStore(
	distributionType string,
	steps int,
	duration int,
	workers int,
	minRate int,
	maxRate int,
) {
	distribution := getDistribution(distributionType, steps, minRate, maxRate)
	targets := getTargets()
	for _, frequency := range distribution {
		rate := vegeta.Rate{Freq: frequency, Per: time.Second}
		attackerFunc := vegeta.Workers(uint64(workers))
		attacker := vegeta.NewAttacker(attackerFunc)
		targeter := vegeta.NewStaticTargeter(targets...)
		res := attacker.Attack(targeter, rate, 10*time.Second, "")
		open := true
		for open {
			_, open = <-res
		}
	}
}
