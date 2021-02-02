package main

type treno struct {
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

type covid []struct {
	Data                               string      `json:"data"`
	Stato                              string      `json:"stato"`
	RicoveratiConSintomi               int         `json:"ricoverati_con_sintomi"`
	TerapiaIntensiva                   int         `json:"terapia_intensiva"`
	TotaleOspedalizzati                int         `json:"totale_ospedalizzati"`
	IsolamentoDomiciliare              int         `json:"isolamento_domiciliare"`
	TotalePositivi                     int         `json:"totale_positivi"`
	VariazioneTotalePositivi           int         `json:"variazione_totale_positivi"`
	NuoviPositivi                      int         `json:"nuovi_positivi"`
	DimessiGuariti                     int         `json:"dimessi_guariti"`
	Deceduti                           int         `json:"deceduti"`
	CasiDaSospettoDiagnostico          interface{} `json:"casi_da_sospetto_diagnostico"`
	CasiDaScreening                    interface{} `json:"casi_da_screening"`
	TotaleCasi                         int         `json:"totale_casi"`
	Tamponi                            int         `json:"tamponi"`
	CasiTestati                        interface{} `json:"casi_testati"`
	Note                               interface{} `json:"note"`
	IngressiTerapiaIntensiva           interface{} `json:"ingressi_terapia_intensiva"`
	NoteTest                           interface{} `json:"note_test"`
	NoteCasi                           interface{} `json:"note_casi"`
	TotalePositiviTestMolecolare       interface{} `json:"totale_positivi_test_molecolare"`
	TotalePositiviTestAntigenicoRapido interface{} `json:"totale_positivi_test_antigenico_rapido"`
	TamponiTestMolecolare              interface{} `json:"tamponi_test_molecolare"`
	TamponiTestAntigenicoRapido        interface{} `json:"tamponi_test_antigenico_rapido"`
}
