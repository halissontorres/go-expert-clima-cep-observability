package model

type ViaCEPResponse struct {
	Localidade string `json:"localidade"`
	Erro       string `json:"erro"`
}

type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

type WeatherInput struct {
	CEP string `json:"cep"`
}

type WeatherOutput struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}
