package main

import (
	"encoding/json"
	"github.com/goodsign/monday"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"io/ioutil"
	"net/http"
	"time"
)

func getCovid() string {

	var (
		covid covid
		p     = message.NewPrinter(language.Italian)
		date  time.Time
	)

	resp, err := http.Get("https://raw.githubusercontent.com/pcm-dpc/COVID-19/master/dati-json/dpc-covid19-ita-andamento-nazionale.json")
	if err != nil {
		return ""
	}

	data, _ := ioutil.ReadAll(resp.Body)
	_ = resp.Body.Close()

	err = json.Unmarshal(data, &covid)
	if err != nil {
		return ""
	}

	date, _ = time.Parse("2006-01-02T15:04:05", covid[len(covid)-1].Data)

	return "Dati del " + monday.Format(date, "2 January 2006", monday.LocaleItIT) + "; Nuovi casi: " + p.Sprintf("%d", covid[len(covid)-1].NuoviPositivi) + "; Numero di tamponi effettuati oggi: " + p.Sprintf("%d", covid[len(covid)-1].Tamponi-covid[len(covid)-2].Tamponi) + "; Numero di morti oggi: " + p.Sprintf("%d", covid[len(covid)-1].Deceduti-covid[len(covid)-2].Deceduti) + "; Totale positivi: " + p.Sprintf("%d", covid[len(covid)-1].TotalePositivi)

}
