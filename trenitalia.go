package main

import (
	"encoding/json"
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

	body, err := ioutil.ReadAll(resp.Body)

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

	res, err := http.Get("http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/andamentoTreno/" + idStazioneTreno)
	if err != nil {
		return ""
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ""
	}

	err = json.Unmarshal(body, &treno)
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

	return "Il treno " + treno.CompTipologiaTreno + ", " + strconv.Itoa(treno.NumeroTreno) + ", di trenitalia, proveniente da " + treno.Origine + " ,e diretto a " + treno.Destinazione + ", delle ore " + ora.Format("15:04") + ", e' in arrivo al binario " + binario + "! Attenzione! Allontanarsi dalla linea gialla! Ferma a: " + stazioni

}
