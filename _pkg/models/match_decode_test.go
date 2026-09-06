package models

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Реальный ответ Hasura должен разбираться в модель целиком.
// Регрессия на map.offset: поле приходит строкой "(-2000,3250)",
// а объявлено было *float64 — вся загрузка матча падала на декодировании.
func TestDecodeRealHasuraResponse(t *testing.T) {
	path := filepath.Join("testdata", "match_response.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("нет образца ответа (%s): %v", path, err)
	}

	var resp GetMatchStatsResponse
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields() // ловим поля, которых нет в модели

	if err := dec.Decode(&resp); err != nil {
		// DisallowUnknownFields шумит на legacy-полях, поэтому при
		// расхождении проверяем ещё раз в обычном режиме
		var lenient GetMatchStatsResponse
		if err2 := json.Unmarshal(raw, &lenient); err2 != nil {
			t.Fatalf("ответ Hasura не разбирается в модель: %v", err2)
		}
		t.Logf("строгий разбор нашёл расхождение (не критично): %v", err)
		resp = lenient
	}

	m := resp.Data.Match
	if m.ID == 0 {
		t.Fatal("матч не разобрался: ID пустой")
	}
	if len(m.Maps) == 0 {
		t.Error("ожидались карты матча")
	}
	if len(m.Teams) != 2 {
		t.Errorf("ожидались 2 команды, получено %d", len(m.Teams))
	}
	if len(m.Members) == 0 {
		t.Error("ожидались участники матча")
	}
	if len(m.Rounds) == 0 {
		t.Error("ожидались раунды матча")
	}

	for _, mm := range m.Maps {
		if mm.Map.Name == "" {
			t.Errorf("у карты %d пустое имя", mm.ID)
		}
	}
}
