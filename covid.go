package main

import (
	"github.com/gocarina/gocsv"
	"github.com/goodsign/monday"
	"net/http"
	"strconv"
	"time"
)

const dataUrl = "https://github.com/pcm-dpc/COVID-19/raw/master/dati-andamento-nazionale/dpc-covid19-ita-andamento-nazionale.csv"

type Covid struct {
	Data                         DateTime `csv:"data"`
	Stato                        string   `csv:"stato"`
	Ricoverati_con_sintomi       int      `csv:"ricoverati_con_sintomi"`
	Terapia_intensiva            int      `csv:"terapia_intensiva"`
	Totale_ospedalizzati         int      `csv:"totale_ospedalizzati"`
	Isolamento_domiciliare       int      `csv:"isolamento_domiciliare"`
	Totale_positivi              int      `csv:"totale_positivi"`
	Variazione_totale_positivi   int      `csv:"variazione_totale_positivi"`
	Nuovi_positivi               int      `csv:"nuovi_positivi"`
	Dimessi_guariti              int      `csv:"dimessi_guariti"`
	Deceduti                     int      `csv:"deceduti"`
	Casi_da_sospetto_diagnostico string   `csv:"casi_da_sospetto_diagnostico"`
	Casi_da_screening            string   `csv:"casi_da_screening"`
	Totale_casi                  int      `csv:"totale_casi"`
	Tamponi                      int      `csv:"tamponi"`
	Casi_testati                 string   `csv:"casi_testati"`
	Note                         string   `csv:"note"`
}

type DateTime struct {
	time.Time
}

// Convert the CSV string as internal date
func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("2006-01-02T15:04:05", csv)
	return err
}

func getCovid() string {
	var covid []*Covid

	resp, err := http.Get(dataUrl)
	if err != nil {
		return ""
	}

	_ = gocsv.Unmarshal(resp.Body, &covid)
	_ = resp.Body.Close()

	return "Dati del " + monday.Format(covid[len(covid)-1].Data.Time, "2 January 2006", monday.LocaleItIT) + ". Totale positivi: " + strconv.Itoa(covid[len(covid)-1].Totale_positivi) + "; Numero di tamponi effettuati oggi: " + strconv.Itoa(covid[len(covid)-1].Tamponi) + "; Numero di morti oggi: " + strconv.Itoa(covid[len(covid)-1].Deceduti) + "; Incremento di casi rispetto a ieri: " + strconv.Itoa(covid[len(covid)-1].Variazione_totale_positivi)

}
