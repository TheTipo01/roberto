package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Treno struct {
	TipoTreno        string      `json:"tipoTreno"`
	Orientamento     interface{} `json:"orientamento"`
	CodiceCliente    int         `json:"codiceCliente"`
	FermateSoppresse interface{} `json:"fermateSoppresse"`
	DataPartenza     interface{} `json:"dataPartenza"`
	Fermate          []struct {
		Orientamento                          interface{} `json:"orientamento"`
		KcNumTreno                            interface{} `json:"kcNumTreno"`
		Stazione                              string      `json:"stazione"`
		ID                                    string      `json:"id"`
		ListaCorrispondenze                   interface{} `json:"listaCorrispondenze"`
		Programmata                           int64       `json:"programmata"`
		ProgrammataZero                       interface{} `json:"programmataZero"`
		Effettiva                             interface{} `json:"effettiva"`
		Ritardo                               int         `json:"ritardo"`
		PartenzaTeoricaZero                   interface{} `json:"partenzaTeoricaZero"`
		ArrivoTeoricoZero                     interface{} `json:"arrivoTeoricoZero"`
		PartenzaTeorica                       int64       `json:"partenza_teorica"`
		ArrivoTeorico                         interface{} `json:"arrivo_teorico"`
		IsNextChanged                         bool        `json:"isNextChanged"`
		PartenzaReale                         interface{} `json:"partenzaReale"`
		ArrivoReale                           interface{} `json:"arrivoReale"`
		RitardoPartenza                       int         `json:"ritardoPartenza"`
		RitardoArrivo                         int         `json:"ritardoArrivo"`
		Progressivo                           int         `json:"progressivo"`
		BinarioEffettivoArrivoCodice          interface{} `json:"binarioEffettivoArrivoCodice"`
		BinarioEffettivoArrivoTipo            interface{} `json:"binarioEffettivoArrivoTipo"`
		BinarioEffettivoArrivoDescrizione     interface{} `json:"binarioEffettivoArrivoDescrizione"`
		BinarioProgrammatoArrivoCodice        interface{} `json:"binarioProgrammatoArrivoCodice"`
		BinarioProgrammatoArrivoDescrizione   interface{} `json:"binarioProgrammatoArrivoDescrizione"`
		BinarioEffettivoPartenzaCodice        interface{} `json:"binarioEffettivoPartenzaCodice"`
		BinarioEffettivoPartenzaTipo          interface{} `json:"binarioEffettivoPartenzaTipo"`
		BinarioEffettivoPartenzaDescrizione   interface{} `json:"binarioEffettivoPartenzaDescrizione"`
		BinarioProgrammatoPartenzaCodice      interface{} `json:"binarioProgrammatoPartenzaCodice"`
		BinarioProgrammatoPartenzaDescrizione string      `json:"binarioProgrammatoPartenzaDescrizione"`
		TipoFermata                           string      `json:"tipoFermata"`
		VisualizzaPrevista                    bool        `json:"visualizzaPrevista"`
		NextChanged                           bool        `json:"nextChanged"`
		NextTrattaType                        int         `json:"nextTrattaType"`
		ActualFermataType                     int         `json:"actualFermataType"`
		MaterialeLabel                        interface{} `json:"materiale_label"`
	} `json:"fermate"`
	Anormalita                interface{} `json:"anormalita"`
	Provvedimenti             interface{} `json:"provvedimenti"`
	Segnalazioni              interface{} `json:"segnalazioni"`
	OraUltimoRilevamento      interface{} `json:"oraUltimoRilevamento"`
	StazioneUltimoRilevamento string      `json:"stazioneUltimoRilevamento"`
	IDDestinazione            string      `json:"idDestinazione"`
	IDOrigine                 string      `json:"idOrigine"`
	CambiNumero               []struct {
		NuovoNumeroTreno string `json:"nuovoNumeroTreno"`
		Stazione         string `json:"stazione"`
	} `json:"cambiNumero"`
	HasProvvedimenti                      bool          `json:"hasProvvedimenti"`
	DescOrientamento                      []string      `json:"descOrientamento"`
	CompOraUltimoRilevamento              string        `json:"compOraUltimoRilevamento"`
	MotivoRitardoPrevalente               interface{}   `json:"motivoRitardoPrevalente"`
	DescrizioneVCO                        string        `json:"descrizioneVCO"`
	MaterialeLabel                        interface{}   `json:"materiale_label"`
	NumeroTreno                           int           `json:"numeroTreno"`
	Categoria                             string        `json:"categoria"`
	CategoriaDescrizione                  interface{}   `json:"categoriaDescrizione"`
	Origine                               string        `json:"origine"`
	CodOrigine                            interface{}   `json:"codOrigine"`
	Destinazione                          string        `json:"destinazione"`
	CodDestinazione                       interface{}   `json:"codDestinazione"`
	OrigineEstera                         interface{}   `json:"origineEstera"`
	DestinazioneEstera                    interface{}   `json:"destinazioneEstera"`
	OraPartenzaEstera                     interface{}   `json:"oraPartenzaEstera"`
	OraArrivoEstera                       interface{}   `json:"oraArrivoEstera"`
	Tratta                                int           `json:"tratta"`
	Regione                               int           `json:"regione"`
	OrigineZero                           string        `json:"origineZero"`
	DestinazioneZero                      string        `json:"destinazioneZero"`
	OrarioPartenzaZero                    int64         `json:"orarioPartenzaZero"`
	OrarioArrivoZero                      int64         `json:"orarioArrivoZero"`
	Circolante                            bool          `json:"circolante"`
	BinarioEffettivoArrivoCodice          interface{}   `json:"binarioEffettivoArrivoCodice"`
	BinarioEffettivoArrivoDescrizione     interface{}   `json:"binarioEffettivoArrivoDescrizione"`
	BinarioEffettivoArrivoTipo            interface{}   `json:"binarioEffettivoArrivoTipo"`
	BinarioProgrammatoArrivoCodice        interface{}   `json:"binarioProgrammatoArrivoCodice"`
	BinarioProgrammatoArrivoDescrizione   interface{}   `json:"binarioProgrammatoArrivoDescrizione"`
	BinarioEffettivoPartenzaCodice        interface{}   `json:"binarioEffettivoPartenzaCodice"`
	BinarioEffettivoPartenzaDescrizione   interface{}   `json:"binarioEffettivoPartenzaDescrizione"`
	BinarioEffettivoPartenzaTipo          interface{}   `json:"binarioEffettivoPartenzaTipo"`
	BinarioProgrammatoPartenzaCodice      interface{}   `json:"binarioProgrammatoPartenzaCodice"`
	BinarioProgrammatoPartenzaDescrizione interface{}   `json:"binarioProgrammatoPartenzaDescrizione"`
	SubTitle                              string        `json:"subTitle"`
	EsisteCorsaZero                       string        `json:"esisteCorsaZero"`
	InStazione                            bool          `json:"inStazione"`
	HaCambiNumero                         bool          `json:"haCambiNumero"`
	NonPartito                            bool          `json:"nonPartito"`
	Provvedimento                         int           `json:"provvedimento"`
	Riprogrammazione                      interface{}   `json:"riprogrammazione"`
	OrarioPartenza                        int64         `json:"orarioPartenza"`
	OrarioArrivo                          int64         `json:"orarioArrivo"`
	StazionePartenza                      interface{}   `json:"stazionePartenza"`
	StazioneArrivo                        interface{}   `json:"stazioneArrivo"`
	StatoTreno                            interface{}   `json:"statoTreno"`
	Corrispondenze                        interface{}   `json:"corrispondenze"`
	Servizi                               []interface{} `json:"servizi"`
	Ritardo                               int           `json:"ritardo"`
	TipoProdotto                          string        `json:"tipoProdotto"`
	CompOrarioPartenzaZeroEffettivo       string        `json:"compOrarioPartenzaZeroEffettivo"`
	CompOrarioArrivoZeroEffettivo         string        `json:"compOrarioArrivoZeroEffettivo"`
	CompOrarioPartenzaZero                string        `json:"compOrarioPartenzaZero"`
	CompOrarioArrivoZero                  string        `json:"compOrarioArrivoZero"`
	CompOrarioArrivo                      string        `json:"compOrarioArrivo"`
	CompOrarioPartenza                    string        `json:"compOrarioPartenza"`
	CompNumeroTreno                       string        `json:"compNumeroTreno"`
	CompOrientamento                      []string      `json:"compOrientamento"`
	CompTipologiaTreno                    string        `json:"compTipologiaTreno"`
	CompClassRitardoTxt                   string        `json:"compClassRitardoTxt"`
	CompClassRitardoLine                  string        `json:"compClassRitardoLine"`
	CompImgRitardo2                       string        `json:"compImgRitardo2"`
	CompImgRitardo                        string        `json:"compImgRitardo"`
	CompRitardo                           []string      `json:"compRitardo"`
	CompRitardoAndamento                  []string      `json:"compRitardoAndamento"`
	CompInStazionePartenza                []string      `json:"compInStazionePartenza"`
	CompInStazioneArrivo                  []string      `json:"compInStazioneArrivo"`
	CompOrarioEffettivoArrivo             string        `json:"compOrarioEffettivoArrivo"`
	CompDurata                            string        `json:"compDurata"`
	CompImgCambiNumerazione               string        `json:"compImgCambiNumerazione"`
}

func ricercaAndGetTreno(idTreno string) string {
	resp, err := http.Get("http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/cercaNumeroTrenoTrenoAutocomplete/" + idTreno)
	if err != nil {
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)

	if strings.TrimSpace(string(body)) == "" {
		return ""
	}

	combinato := strings.Split(strings.Split(string(body), "|")[1], "-")

	_ = resp.Body.Close()

	return getTreno(strings.TrimSpace(combinato[1]) + "/" + combinato[0])
}

func getTreno(idStazioneTreno string) string {
	pls := true
	var stazioni, binario string
	var ritardo int
	var ora time.Time
	rand.Seed(time.Now().Unix())
	url := "http://www.viaggiatreno.it/viaggiatrenonew/resteasy/viaggiatreno/andamentoTreno/" + idStazioneTreno

	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("User-Agent", "spacecount-tutorial")

	res, err := client.Do(req)
	if err != nil {
		return ""
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ""
	}

	treno := Treno{}

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
