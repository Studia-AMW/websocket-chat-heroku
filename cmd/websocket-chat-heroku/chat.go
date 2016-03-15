package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Wiadomość wysłana przez klienta korzystającego z JavaScript
type message struct {
	Handle string `json:"handle"`
	Text   string `json:"text"`
}

// Walidacja wiadomości
func validateMessage(data []byte) (message, error) {
	var msg message

	if err := json.Unmarshal(data, &msg); err != nil {
		return msg, err
	}

	if msg.Handle == "" && msg.Text == "" {
		return msg, fmt.Errorf("Wiadomość nie zawiera tekstu lub uchwytu")
	}

	return msg, nil
}

// handleWebsocket connection. Update to
func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Niedozwolona metoda", 405)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.WithField("err", err).Println("Przejście do Websocket")
		http.Error(w, "Error Upgrading to websockets", 400)
		return
	}

	id := rr.register(ws)

	for {
		mt, data, err := ws.ReadMessage()
		ctx := log.Fields{"mt": mt, "data": data, "err": err}
		if err != nil {
			if err == io.EOF {
				log.WithFields(ctx).Info("Połączenie websocket zamknięte")
			} else {
				log.WithFields(ctx).Error("Błąd odczytu wiadomości websocket")
			}
			break
		}
		switch mt {
		case websocket.TextMessage:
			msg, err := validateMessage(data)
			if err != nil {
				ctx["msg"] = msg
				ctx["err"] = err
				log.WithFields(ctx).Error("Niepoprawna wiadomość")
				break
			}
			rw.publish(data)
		default:
			log.WithFields(ctx).Warning("Nieznana wiadomość")
		}
	}

	rr.deRegister(id)

	ws.WriteMessage(websocket.CloseMessage, []byte{})
}
