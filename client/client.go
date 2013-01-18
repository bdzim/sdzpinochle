// sdzpinochle-client project main.go
package main

import (
	"encoding/json"
	"fmt"
	sdz "github.com/mzimmerman/sdzpinochle"
	"net"
)

func send(enc *json.Encoder, action *sdz.Action) {
	err := enc.Encode(action)
	if err != nil {
		sdz.Log("Error sending - %v", err)
	} else {
		//sdz.Log("Action sent to server = %v", action)
	}

}

func main() {
	conn, err := net.Dial("tcp", "localhost:1201")
	var playerid int
	var hand *sdz.Hand
	var bidAmount int
	var trump sdz.Suit
	if err != nil {
		sdz.Log("Error - %v", err)
		return
	}
	defer conn.Close()
	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	for {
		var action sdz.Action
		err := dec.Decode(&action)
		if err != nil {
			sdz.Log("Error decoding - %v", err)
			return
		}
		//sdz.Log("Action received from server = %v", action)
		switch action.Type {
		case "Bid":
			if action.Playerid == playerid {
				sdz.Log("How much would you like to bid?:")
				fmt.Scan(&bidAmount)
				send(enc, sdz.CreateBid(bidAmount, playerid))
			} else {
				// received someone else's bid value'
				sdz.Log("Player #%d bid %d", action.Playerid, action.Bid)
			}
		case "Play":
			if action.Playerid == playerid {
				var card sdz.Card
				sdz.Log("Your turn, in your hand is %s - what would you like to play? Trump is %s:", hand, trump)
				fmt.Scan(&card)
				//sdz.Log("Received input %s", card)
				if hand.Remove(card) {
					send(enc, sdz.CreatePlay(card, playerid))
				}
			} else {
				sdz.Log("Player %d played card %s", action.Playerid, action.PlayedCard)
				// received someone else's play'
			}
		case "Trump":
			if action.Playerid == playerid {
				sdz.Log("What would you like to make trump?")
				fmt.Scan(&trump)
				send(enc, sdz.CreateTrump(trump, playerid))
			} else {
				sdz.Log("Player %d says trump is %s", action.Playerid, action.Trump)
				trump = action.Trump
			}
		case "Throwin":
			sdz.Log("Player %d threw in", action.Playerid)
		case "Deal":
			playerid = action.Playerid
			hand = &action.Hand
			sdz.Log("Your hand is - %s", hand)
		case "Meld":
			sdz.Log("Player %d is melding %s for %d points", action.Playerid, action.Hand, action.Amount)
		case "Message":
			sdz.Log(action.Message)
		case "Hello":
			var response string
			fmt.Scan(&response)
			send(enc, sdz.CreateHello(response))
		case "Game":
			var option int
			fmt.Scan(&option)
			send(enc, sdz.CreateGame(option))
		default:
			sdz.Log("Received an action I didn't understand - %v", action)
		}

	}
}
