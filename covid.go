package main

import (
	"github.com/gocarina/gocsv"
	"github.com/goodsign/monday"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"net/http"
	"time"
)

type Covid struct {
	Data                      DateTime `csv:"data"`
	Stato                     string   `csv:"stato"`
	RicoveratiConSintomi      int      `csv:"ricoverati_con_sintomi"`
	TerapiaIntensiva          int      `csv:"terapia_intensiva"`
	TotaleOspedalizzati       int      `csv:"totale_ospedalizzati"`
	IsolamentoDomiciliare     int      `csv:"isolamento_domiciliare"`
	TotalePositivi            int      `csv:"totale_positivi"`
	VariazioneTotalePositivi  int      `csv:"variazione_totale_positivi"`
	NuoviPositivi             int      `csv:"nuovi_positivi"`
	DimessiGuariti            int      `csv:"dimessi_guariti"`
	Deceduti                  int      `csv:"deceduti"`
	CasiDaSospettoDiagnostico string   `csv:"casi_da_sospetto_diagnostico"`
	CasiDaScreening           string   `csv:"casi_da_screening"`
	TotaleCasi                int      `csv:"totale_casi"`
	Tamponi                   int      `csv:"tamponi"`
	CasiTestati               string   `csv:"casi_testati"`
	Note                      string   `csv:"note"`
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
	p := message.NewPrinter(language.Italian)

	resp, err := http.Get("https://github.com/pcm-dpc/COVID-19/raw/master/dati-andamento-nazionale/dpc-covid19-ita-andamento-nazionale.csv")
	if err != nil {
		return ""
	}

	_ = gocsv.Unmarshal(resp.Body, &covid)
	_ = resp.Body.Close()

	return "Dati del " + monday.Format(covid[len(covid)-1].Data.Time, "2 January 2006", monday.LocaleItIT) + "; Nuovi casi: " + p.Sprintf("%d", covid[len(covid)-1].NuoviPositivi) + "; Numero di tamponi effettuati oggi: " + p.Sprintf("%d", (covid[len(covid)-1].Tamponi-covid[len(covid)-2].Tamponi)) + "; Numero di morti oggi: " + p.Sprintf("%d", (covid[len(covid)-1].Deceduti-covid[len(covid)-2].Deceduti)) + "; Totale positivi: " + p.Sprintf("%d", covid[len(covid)-1].TotalePositivi)

}
