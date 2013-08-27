// pinochle.go
package sdzpinochle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"time"
)

func Log(m string, v ...interface{}) {
	fmt.Printf(m+"\n", v...)
}

const (
	Ace         = iota
	Ten         = iota
	King        = iota
	Queen       = iota
	Jack        = iota
	Nine        = iota
	NAFace      = -1
	acearound   = 10
	kingaround  = 8
	queenaround = 6
	jackaround  = 4
	debugLog    = false
	AllCards    = 24
)

const (
	Spades   = iota
	Hearts   = iota
	Clubs    = iota
	Diamonds = iota
	NASuit   = -1
)

const (
	AS     = iota
	TS     = iota
	KS     = iota
	QS     = iota
	JS     = iota
	NS     = iota
	AH     = iota
	TH     = iota
	KH     = iota
	QH     = iota
	JH     = iota
	NH     = iota
	AC     = iota
	TC     = iota
	KC     = iota
	QC     = iota
	JC     = iota
	NC     = iota
	AD     = iota
	TD     = iota
	KD     = iota
	QD     = iota
	JD     = iota
	ND     = iota
	NACard = -1
)

var Faces [6]Face
var Suits [4]Suit

func init() {
	rand.Seed(time.Now().UnixNano())
	Faces = [6]Face{Ace, Ten, King, Queen, Jack, Nine}
	Suits = [4]Suit{Spades, Hearts, Clubs, Diamonds}
}

type Card int // an integer representation of the card
type Suit int
type Face int

type Deck [48]Card
type Hand []Card

func CreateCard(suit Suit, face Face) Card {
	return Card(int(suit)*6 + int(face))
}

func (c Card) String() string {
	if c == NACard {
		return "NA"
	}
	return c.Face().String() + c.Suit().String()
}

func (a Card) Beats(b Card, trump Suit) bool {
	// a is the challenging card
	if b == NACard {
		return true
	}
	switch {
	case a.Suit() == b.Suit():
		return a < b
	case a.Suit() == trump:
		return true
	}
	return false
}

func (c Card) Counter() bool {
	return c.Face() == Ace || c.Face() == Ten || c.Face() == King
}

func (c Card) Suit() Suit {
	if c == NACard {
		return NASuit
	}
	return Suit(int(c) / 6)
}

func (c Card) Face() Face {
	if c == NACard {
		return NAFace
	}
	return Face(int(c) % 6)
}

func (d *Deck) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d *Deck) Shuffle() {
	//	http://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle#The_modern_algorithm
	for i := len(d) - 1; i >= 1; i-- {
		if j := rand.Intn(i); i != j {
			d.Swap(i, j)
		}
	}
}

func (h Hand) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Hand{")
	for x := range h {
		buffer.WriteString(h[x].String())
		buffer.WriteString(" ")
	}
	buffer.WriteString("}")
	return buffer.String()
}

func (h Hand) Len() int {
	return len(h)
}

func (h Hand) Less(i, j int) bool {
	if h[i].Suit() == h[j].Suit() {
		return h[i].Face().Less(h[j].Face())
	}
	return h[i].Suit().Less(h[j].Suit())
}

func (a Face) Less(b Face) bool {
	return a < b
}

func (a Suit) String() string {
	switch a {
	case NASuit:
		return "~"
	case Diamonds:
		return "D"
	case Spades:
		return "S"
	case Hearts:
		return "H"
	case Clubs:
		return "C"
	}
	panic(fmt.Sprintf("Error finding suit for %d", a))
}

func (a Face) String() string {
	switch a {
	case Nine:
		return "9"
	case Jack:
		return "J"
	case Queen:
		return "Q"
	case King:
		return "K"
	case Ten:
		return "T"
	case Ace:
		return "A"
	}
	panic(fmt.Sprintf("Error finding face for %d", int(a)))
}

func (a Suit) Less(b Suit) bool { // only for sorting the suits for display in the hand
	return a > b
}

func (h Hand) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *Hand) Shuffle() {
	//	http://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle#The_modern_algorithm
	for i := len(*h) - 1; i >= 1; i-- {
		if j := rand.Intn(i); i != j {
			h.Swap(i, j)
		}
	}
}

func (d Deck) Deal() (hands []Hand) {
	hands = make([]Hand, 4)
	for x := 0; x < 4; x++ {
		hands[x] = make([]Card, 12)
	}
	for y := 0; y < 12; y++ {
		for x := 0; x < 4; x++ {
			hands[x][y] = d[y*4+x]
		}
	}
	return
}

func CreateDeck() (deck Deck) {
	for x := 0; x < len(deck); x++ {
		deck[x] = Card(x % AllCards)
	}
	return
}

type Action struct {
	Type                    string
	Playerid                int
	Bid                     int
	PlayedCard, WinningCard Card
	Lead, Trump             Suit
	Amount                  int
	Message                 string
	Hand                    Hand
	TableId                 int64
	GameOver, Win           bool
	Score                   []int
	Dealer                  int
	WinningPlayer           int
}

func (action *Action) String() string {
	data, _ := action.MarshalJSON()
	return string(data)
}

func (action *Action) MarshalJSON() ([]byte, error) {
	data := make(map[string]interface{})
	typ := reflect.TypeOf(*action)
	val := reflect.ValueOf(*action)
	count := typ.NumField()
	for x := 0; x < count; x++ {
		switch {
		case typ.Field(x).Name == "Playerid":
			if action.Type == "Hello" || action.Type == "Score" || action.Type == "Message" || action.Type == "Game" {
				// don't include playerid', it's not relevant'
			} else {
				data["Playerid"] = action.Playerid
			}
		case typ.Field(x).Name == "WinningPlayer" && action.Type == "Play":
			data["WinningPlayer"] = action.WinningPlayer
		case typ.Field(x).Name == "Amount" && action.Type == "Bid":
			data["Amount"] = action.Amount
		case typ.Field(x).Name == "Win" && action.GameOver:
			data["Win"] = action.Win
		case typ.Field(x).Name == "GameOver" && action.Type == "Score":
			data["GameOver"] = action.GameOver
		case typ.Field(x).Name == "Dealer" && action.Type == "Deal":
			data["Dealer"] = action.Dealer
		case reflect.DeepEqual(val.Field(x).Interface(), reflect.New(typ.Field(x).Type).Elem().Interface()):
		default:
			data[typ.Field(x).Name] = val.Field(x).Interface()
		}
	}
	return json.Marshal(data)
}

func CreateName() *Action {
	return &Action{Type: "Name"}
}

func CreateSit(tableid int64) *Action {
	return &Action{Type: "Sit", TableId: tableid}
}

func CreateMessage(m string) *Action {
	return &Action{Type: "Message", Message: m}
}

func CreateBid(bid, playerid int) *Action {
	return &Action{Type: "Bid", Bid: bid, Playerid: playerid}
}

func CreatePlayRequest(winning Card, lead, trump Suit, playerid int, hand *Hand) *Action {
	return &Action{Type: "Play", WinningCard: winning, Lead: lead, Trump: trump, Playerid: playerid, Hand: *hand}
}

func CreatePlay(card Card, playerid int) *Action {
	return &Action{Type: "Play", PlayedCard: card, Playerid: playerid}
}

func CreateTrump(trump Suit, playerid int) *Action {
	return &Action{Type: "Trump", Trump: trump, Playerid: playerid}
}

func CreateTrick(winningPlayer int) *Action {
	return &Action{Type: "Trick", Playerid: winningPlayer}
}

func CreateThrowin(playerid int) *Action {
	return &Action{Type: "Throwin", Playerid: playerid}
}

func CreateMeld(hand Hand, amount, playerid int) *Action {
	return &Action{Type: "Meld", Hand: hand, Amount: amount, Playerid: playerid}
}

func CreateDisconnect(playerid int) *Action {
	return &Action{Type: "Disconnect", Playerid: playerid}
}

func CreateDeal(hand Hand, playerid, dealer int) *Action {
	return &Action{Type: "Deal", Hand: hand, Playerid: playerid, Dealer: dealer}
}

func CreateScore(score []int, gameOver, win bool) *Action {
	return &Action{Type: "Score", Score: score, Win: win, GameOver: gameOver}
}

type PlayerImpl struct {
	Playerid int
}

func (p PlayerImpl) PlayerID() int {
	return p.Playerid
}

func (p PlayerImpl) Team() int {
	return p.Playerid % 2
}

func (p PlayerImpl) IsPartner(player int) bool {
	return p.Playerid%2 == player%2
}

// Used to determine if the leader of the trick made a valid play
func IsCardInHand(card Card, hand Hand) bool {
	for _, hc := range hand {
		if hc == card {
			return true
		}
	}
	return false
}

// playedCard, winningCard Card, leadSuit Suit, hand Hand, trump Suit
func ValidPlay(playedCard, winningCard Card, leadSuit Suit, hand *Hand, trump Suit) bool {
	if winningCard == NACard || leadSuit == NASuit {
		return true
	}
	// hand is sorted
	// 1 - Have to follow suit
	// 2 - Can't follow suit, play trump
	// 3 - Have to win
	canFollow := false
	hasTrump := false
	canWin := false
	hasCard := false
	for _, card := range *hand {
		if card.Suit() == leadSuit {
			canFollow = true
		}
		if card.Suit() == trump {
			hasTrump = true
		}
		if card == playedCard {
			hasCard = true
		}
	}
	if !hasCard { // you don't have the card in your hand, not allowed to play it, cheater!
		return false
	}
	if winningCard == NACard { // nothing to follow so far, so you win!
		return true
	}

	// have to loop again because we can't set canWin to true if we're playing trump but we can follow a non-trump suit
	for _, card := range *hand {
		if canFollow && leadSuit != trump && card.Suit() == trump {
			continue
		}
		if card.Beats(winningCard, trump) {
			canWin = true
			break
		}
	}
	if canFollow {
		if playedCard.Suit() != leadSuit {
			return false
		} else if canWin { // we're following suit
			return playedCard.Beats(winningCard, trump)
		} else { // we're following suit and we can't win'
			return true
		}
	} else if hasTrump {
		if playedCard.Suit() != trump {
			return false
		} else if canWin { // we're playing trump
			return playedCard.Beats(winningCard, trump)
		} else { // we're playing trump but we can't win
			return true
		}
	} // else { // we can't follow suit and we don't have trump - anything's legal
	return true
}

func (h *Hand) Contains(card Card) bool {
	for _, c := range *h {
		if c == card {
			return true
		}
	}
	return false
}

func (h *Hand) Remove(card Card) bool {
	for x := range *h {
		if (*h)[x] == card {
			//temp := append((*h)[:x], (*h)[x+1:]...)
			//h = &temp
			*h = append((*h)[:x], (*h)[x+1:]...)
			return true
		}
	}
	return false
}

func (h Hand) CountSuit(suit Suit) (count int) {
	for _, card := range h {
		if card.Suit() == suit {
			count++
		}
	}
	return
}

func (h Hand) Count() (cards map[Card]int) {
	cards = make(map[Card]int)
	for _, face := range Faces {
		for _, suit := range Suits {
			cards[CreateCard(suit, face)] = 0
		}
	}
	for x := 0; x < len(h); x++ {
		cards[h[x]]++
	}
	return
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h Hand) Meld(trump Suit) (meld int, result Hand) {
	// hand does not have to be sorted
	count := h.Count()
	if debugLog {
		fmt.Printf("Count is %v\n", count)
	}
	show := make(map[Card]int)
	around := make(map[Face]int)
	for _, value := range Faces {
		around[value] = 2
	}
	//	fmt.Printf("AroundBefore = %v\n", around)
	for _, suit := range Suits { // look through each suit
		switch { // straights & marriages
		case trump == suit:
			if debugLog {
				fmt.Printf("Scoring %d nine(s) in trump %s\n", count[CreateCard(suit, Nine)], trump)
			}
			meld += count[CreateCard(suit, Nine)] // 9s in trump
			show[CreateCard(suit, Nine)] = count[CreateCard(suit, Nine)]
			switch {
			// double straight
			case count[CreateCard(suit, Ace)] == 2 && count[CreateCard(suit, Ten)] == 2 && count[CreateCard(suit, King)] == 2 && count[CreateCard(suit, Queen)] == 2 && count[CreateCard(suit, Jack)] == 2:
				meld += 150
				for _, face := range Faces {
					show[CreateCard(suit, face)] = 2
				}
				if debugLog {
					fmt.Println("DoubleStraight")
				}
			// single straight
			case count[CreateCard(suit, Ace)] >= 1 && count[CreateCard(suit, Ten)] >= 1 && count[CreateCard(suit, King)] >= 1 && count[CreateCard(suit, Queen)] >= 1 && count[CreateCard(suit, Jack)] >= 1:
				for _, face := range []Face{Ace, Ten, King, Queen, Jack} {
					show[CreateCard(suit, face)] = max(show[CreateCard(suit, face)], 1)
				}
				if count[CreateCard(suit, King)] == 2 && count[CreateCard(suit, Queen)] == 2 {
					show[CreateCard(suit, King)] = 2
					show[CreateCard(suit, Queen)] = 2
					meld += 19
					if debugLog {
						fmt.Println("SingleStraightWithExtraMarriage")
					}
				} else {
					if debugLog {
						fmt.Println("SingleStraight")
					}
					meld += 15
				}
			case count[CreateCard(suit, King)] == 2 && count[CreateCard(suit, Queen)] == 2:
				meld += 8
				show[CreateCard(suit, King)] = 2
				show[CreateCard(suit, Queen)] = 2
				if debugLog {
					fmt.Println("DoubleMarriageInTrump")
				}
			case count[CreateCard(suit, King)] >= 1 && count[CreateCard(suit, Queen)] >= 1:
				meld += 4
				show[CreateCard(suit, King)] = max(show[CreateCard(suit, King)], 1)
				show[CreateCard(suit, Queen)] = max(show[CreateCard(suit, Queen)], 1)
				if debugLog {
					fmt.Println("SingleMarriageInTrump")
				}
			}
		case count[CreateCard(suit, King)] == 2 && count[CreateCard(suit, Queen)] == 2:
			show[CreateCard(suit, King)] = 2
			show[CreateCard(suit, Queen)] = 2
			meld += 4
			if debugLog {
				fmt.Println("DoubleMarriage")
			}
		case count[CreateCard(suit, King)] >= 1 && count[CreateCard(suit, Queen)] >= 1:
			show[CreateCard(suit, King)] = max(show[CreateCard(suit, King)], 1)
			show[CreateCard(suit, Queen)] = max(show[CreateCard(suit, Queen)], 1)
			if debugLog {
				fmt.Println("SingleMarriage")
			}
			meld += 2
		}
		for _, face := range Faces { // looking for "around" meld
			//						fmt.Printf("Looking for %d in suit %d\n", value, suit)
			around[face] = min(count[CreateCard(suit, face)], around[face])
		}
	}
	for _, face := range []Face{Ace, King, Queen, Jack} {
		if around[face] > 0 {
			var worth int
			switch face {
			case Ace:
				worth = acearound
			case King:
				worth = kingaround
			case Queen:
				worth = queenaround
			case Jack:
				worth = jackaround
			}
			if around[face] == 2 {
				worth *= 10
			}
			for _, suit := range Suits {
				show[CreateCard(suit, face)] = max(show[CreateCard(suit, face)], around[face])
			}
			meld += worth
			if debugLog {
				fmt.Printf("Around-%d\n", worth)
			}
		}
	}
	switch { // pinochle
	case count[CreateCard(Diamonds, Jack)] == 2 && count[CreateCard(Spades, Queen)] == 2:
		meld += 30
		show[CreateCard(Spades, Queen)] = 2
		show[CreateCard(Diamonds, Jack)] = 2
		if debugLog {
			fmt.Println("DoubleNochle")
		}
	case count[CreateCard(Diamonds, Jack)] >= 1 && count[CreateCard(Spades, Queen)] >= 1:
		meld += 4
		show[CreateCard(Diamonds, Jack)] = max(show[CreateCard(Diamonds, Jack)], 1)
		show[CreateCard(Spades, Queen)] = max(show[CreateCard(Spades, Queen)], 1)
		if debugLog {
			fmt.Println("Nochle")
		}
	}
	result = make([]Card, 0, 12)
	for card, amount := range show {
		for {
			if amount > 0 {
				result = append(result, card)
				amount--
			} else {
				break
			}
		}
	}
	sort.Sort(result)
	return
}
