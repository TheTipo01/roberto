package main

import (
	"github.com/goccy/go-json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Search where the given trainID starts
func searchAndGetTrain(trainID string) string {

	resp, err := http.Get("http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/cercaNumeroTrenoTrenoAutocomplete/" + trainID)
	if err != nil {
		return ""
	}

	body, _ := ioutil.ReadAll(resp.Body)

	out := string(body)
	_ = resp.Body.Close()

	if strings.TrimSpace(out) == "" {
		return ""
	}

	foo := strings.Split(out, "|")

	if len(foo) >= 1 {
		combinato := strings.Split(foo[1], "-")

		if len(combinato) >= 1 {
			return getTrain(strings.TrimSpace(combinato[1]) + "/" + combinato[0])
		}
	}

	return ""
}

// Returns text for a given train
func getTrain(idStazioneTreno string) string {

	var (
		pls               = true
		stazioni, binario string
		ritardo           int
		ora               time.Time
		treno             = treno{}
	)

	res, err := http.Get("http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/andamentoTreno/" + idStazioneTreno + "/" + midnight())
	if err != nil {
		return ""
	}

	err = json.NewDecoder(res.Body).Decode(&treno)
	_ = res.Body.Close()
	if err != nil {
		return ""
	}

	for _, stazione := range treno.Fermate {
		if stazione.Stazione != treno.Origine && stazione.Stazione != treno.Destinazione && !pls {
			stazioni += stazione.Stazione + ","
		}

		if pls && stazione.Effettiva == nil {
			binario = stazione.BinarioProgrammatoPartenzaDescrizione
			ora = time.Unix(stazione.Programmata/1000, 0)
			ora = ora.Add(time.Minute * time.Duration(ritardo))
			pls = false
		} else {
			ritardo = stazione.Ritardo
		}

	}

	stazioni = strings.TrimSuffix(stazioni, ",") + "."

	if stazioni != "." {
		return "Il treno " + treno.CompTipologiaTreno + ", " + strconv.Itoa(treno.NumeroTreno) + ", di trenitalia, proveniente da " + treno.Origine + " ,e diretto a " + treno.Destinazione + ", delle ore " + ora.Format("15:04") + ", e' in arrivo al binario " + binario + "! Attenzione! Allontanarsi dalla linea gialla! Ferma a: " + stazioni
	}

	return "Il treno " + treno.CompTipologiaTreno + ", " + strconv.Itoa(treno.NumeroTreno) + ", di trenitalia, proveniente da " + treno.Origine + " ,e diretto a " + treno.Destinazione + ", delle ore " + ora.Format("15:04") + ", e' in arrivo al binario " + binario + "! Attenzione! Allontanarsi dalla linea gialla!"

}

// Returns strange value (seems to be the midnight of the current day multiplied by 1000) that the API needs at the end for some calls. Don't ask, I didn't.
func midnight() string {
	t := time.Now()
	return strconv.FormatInt(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local).Unix()*1000, 10)
}
