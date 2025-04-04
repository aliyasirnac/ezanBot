package thirdparty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type ezanAPI struct {
	BaseUrl string
}

type EzanRequest struct {
	IlceId int
}

type EzanResponse struct {
	EzanBody []EzanBody
}

type EzanBody struct {
	MiladiTarihKisa string `json:"MiladiTarihKisa"`
	Imsak           string `json:"Imsak"`
	Gunes           string `json:"Gunes"`
	Ogle            string `json:"Ogle"`
	Ikindi          string `json:"Ikindi"`
	Aksam           string `json:"Aksam"`
	Yatsi           string `json:"Yatsi"`
}

func NewEzan(baseURL string) ThirdParty[EzanRequest, EzanResponse] {
	return &ezanAPI{BaseUrl: baseURL}
}

func (e ezanAPI) Handler(ctx context.Context, r *EzanRequest) (*EzanResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	url := fmt.Sprintf("%s/vakitler/%d", e.BaseUrl, r.IlceId)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		zap.L().Error("request not created", zap.Error(err))
		return nil, fmt.Errorf("istek oluşturulamadı: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("istek zaman aşımına uğradı")
		}
		return nil, fmt.Errorf("istek başarısız: %w", err)
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cevap okunamadı: %w", err)
	}

	var ezanResponse []EzanBody
	if err := json.Unmarshal(bodyText, &ezanResponse); err != nil {
		return nil, fmt.Errorf("JSON parse hatası: %w", err)
	}
	zap.L().Info("Ezan Fetch Successful", zap.Any("ezanRes", len(ezanResponse)))
	return &EzanResponse{EzanBody: ezanResponse}, nil
}
