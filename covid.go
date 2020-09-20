package main

import (
	"github.com/gocarina/gocsv"
	"github.com/goodsign/monday"
	"net/http"
	"strconv"
	"time"
)

type covid struct {
	data                      dateTime `csv:"data"`
	stato                     string   `csv:"stato"`
	ricoveratiConSintomi      int      `csv:"ricoverati_con_sintomi"`
	terapiaIntensiva          int      `csv:"terapia_intensiva"`
	totaleOspedalizzati       int      `csv:"totale_ospedalizzati"`
	isolamentoDomiciliare     int      `csv:"isolamento_domiciliare"`
	totalePositivi            int      `csv:"totale_positivi"`
	variazioneTotalePositivi  int      `csv:"variazione_totale_positivi"`
	nuoviPositivi             int      `csv:"nuovi_positivi"`
	dimessiGuariti            int      `csv:"dimessi_guariti"`
	deceduti                  int      `csv:"deceduti"`
	casiDaSospettoDiagnostico string   `csv:"casi_da_sospetto_diagnostico"`
	casiDaScreening           string   `csv:"casi_da_screening"`
	totaleCasi                int      `csv:"totale_casi"`
	tamponi                   int      `csv:"tamponi"`
	casiTestati               string   `csv:"casi_testati"`
	note                      string   `csv:"note"`
}

type dateTime struct {
	time.Time
}

// Convert the CSV string as internal date
func (date *dateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("2006-01-02T15:04:05", csv)
	return err
}

func getCovid() string {
	var covid []*covid

	resp, err := http.Get("https://github.com/pcm-dpc/COVID-19/raw/master/dati-andamento-nazionale/dpc-covid19-ita-andamento-nazionale.csv")
	if err != nil {
		return ""
	}

	_ = gocsv.Unmarshal(resp.Body, &covid)
	_ = resp.Body.Close()

	return "Dati del " + monday.Format(covid[len(covid)-1].data.Time, "2 January 2006", monday.LocaleItIT) + ". Totale positivi: " + strconv.Itoa(covid[len(covid)-1].totalePositivi) + "; Numero di tamponi effettuati oggi: " + strconv.Itoa(covid[len(covid)-1].tamponi-covid[len(covid)-2].tamponi) + "; Numero di morti oggi: " + strconv.Itoa(covid[len(covid)-1].deceduti-covid[len(covid)-2].deceduti) + "; Incremento di casi rispetto a ieri: " + strconv.Itoa(covid[len(covid)-1].variazioneTotalePositivi)
}
