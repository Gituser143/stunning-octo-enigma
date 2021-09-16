package load

import (
	"fmt"
	"net/url"

	vegeta "github.com/tsenart/vegeta/lib"
)

// checkStatus returns a vegeta target for the status page of teaStore
func (sc *StressClient) checkStatus() vegeta.Target {
	url := url.URL{
		Scheme: sc.scheme,
		Host:   sc.host,
		Path:   "tools.descartes.teastore.webui/status",
	}

	target := vegeta.Target{
		URL:    url.String(),
		Method: "GET",
	}

	return target
}

// checkLogin returns a vegeta target for the status page of teaStore
func (sc *StressClient) checkLogin(username string, password string) vegeta.Target {
	url := url.URL{
		Scheme: sc.scheme,
		Host:   sc.host,
		Path:   "tools.descartes.teastore.webui/loginAction",
	}

	query := url.Query()
	query.Add("username", username)
	query.Add("password", password)

	url.RawQuery = query.Encode()

	target := vegeta.Target{
		URL:    url.String(),
		Method: "GET",
	}

	return target
}

func (sc *StressClient) checkSuccessfulLogin() vegeta.Target {
	return sc.checkLogin("user2", "password")
}

func (sc *StressClient) checkFailedLogin() vegeta.Target {
	return sc.checkLogin("testuser", "password")
}

func (sc *StressClient) getCategoryEndpoints() []vegeta.Target {
	var targets []vegeta.Target
	categoryIDs := []string{"2", "3", "4", "5", "6"}
	pageNumbers := []string{"1", "2", "3", "4", "5"}

	for _, categoryID := range categoryIDs {
		for _, pageNumber := range pageNumbers {
			url := url.URL{
				Scheme: sc.scheme,
				Host:   sc.host,
				Path:   "tools.descartes.teastore.webui/category",
			}

			q := url.Query()

			q.Add("category", categoryID)
			q.Add("page", pageNumber)

			target := vegeta.Target{
				URL:    url.String(),
				Method: "GET",
			}
			targets = append(targets, target)
		}
	}
	return targets
}

func (sc *StressClient) getProductEndpoints() []vegeta.Target {
	var targets []vegeta.Target
	for productID := 100; productID < 150; productID++ {

		url1 := url.URL{
			Scheme: sc.scheme,
			Host:   sc.host,
			Path:   "tools.descartes.teastore.webui/cartAction",
		}
		q1 := url1.Query()
		q1.Add("addToCart", "")
		q1.Add("productid", fmt.Sprintf("%d", productID))
		url1.RawQuery = q1.Encode()

		url2 := url.URL{
			Scheme: sc.scheme,
			Host:   sc.host,
			Path:   "tools.descartes.teastore.webui/cartAction",
		}
		q2 := url1.Query()
		q2.Add("productid", fmt.Sprintf("%d", productID))
		url2.RawQuery = q2.Encode()

		target1 := vegeta.Target{
			URL:    url1.String(),
			Method: "POST",
		}

		target2 := vegeta.Target{
			URL:    url2.String(),
			Method: "GET",
		}

		targets = append(targets, target1)
		targets = append(targets, target2)
	}

	return targets
}

// GetTeaStoreTargets returns vegeta targets for the TeaStore application
func (sc *StressClient) GetTeaStoreTargets() []vegeta.Target {
	var targets []vegeta.Target

	targets = append(targets, sc.checkStatus())
	targets = append(targets, sc.checkFailedLogin())
	targets = append(targets, sc.checkSuccessfulLogin())
	targets = append(targets, sc.getCategoryEndpoints()...)
	targets = append(targets, sc.getProductEndpoints()...)

	return targets
}
