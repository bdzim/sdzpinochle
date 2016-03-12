package main

import (
	"fmt"
	sdz "github.com/mzimmerman/sdzpinochle"
	ai "github.com/mzimmerman/sdzpinochle/ai"
	"runtime"
	"sort"
	"time"
)

const (
	winningScore      int  = 120
	giveUpScore       int  = -500
	numberOfTricks    int  = 12
	debugLog          bool = false
	matchesToSimulate int  = 1000
)

var oponents = make(chan Oponents)
var results = make(chan Result)
var numberOfMatchRunners = runtime.NumCPU()

type Oponents struct {
	player1 ai.Player
	player2 ai.Player
}

type NamedBid struct {
	Name string
	Bid  ai.BiddingStrategy
}

type NamedPlay struct {
	Name string
	Play ai.PlayingStrategy
}

type Result struct {
	playerOneWins int
	playerTwoWins int
}

func main() {
	startTime := time.Now()
	createMatchRunners()
	matchesSimulated := 0
	players := createPlayers()

	player1 := players[0]
	for _, player2 := range players[1:] {
		matchesSimulated += matchesToSimulate
		fmt.Printf("%v vs %v\n", player1.Name, player2.Name)

		var win1, win2 int
		for x := 0; x < numberOfMatchRunners; x++ {
			oponents <- Oponents{player1, player2}
		}
		for x := 0; x < numberOfMatchRunners; x++ {
			result := <-results
			win1 += result.playerOneWins
			win2 += result.playerTwoWins
		}

		fmt.Printf("%v %v wins - %v %v wins\n", player1.Name, win1, player2.Name, win2)
		if win2 > win1 {
			// Winner stays
			player1 = player2
		}
	}
	elapsedSeconds := time.Since(startTime).Seconds()
	fmt.Printf("%v is the champ!\n", player1.Name)
	fmt.Printf(
		"%v matches simulated in %.f2 seconds\n",
		matchesSimulated,
		elapsedSeconds,
	)
	fmt.Printf("%.2f matches simulated per second.\n", float64(matchesSimulated)/elapsedSeconds)
}

func createPlayers() []ai.Player {
	bidding_strategies := []NamedBid{
		NamedBid{"NeverBid", ai.NeverBid},
		NamedBid{"MostMeld", ai.ChooseSuitWithMostMeld},
	}
	for x := 16; x <= 18; x++ {
		bidding_strategies = append(
			bidding_strategies,
			NamedBid{fmt.Sprintf("MostMeldPlus%v", x), ai.MostMeldPlusX(x)},
		)
	}
	playing_strategies := []NamedPlay{
		NamedPlay{"PlayHighest", ai.PlayHighest},
		NamedPlay{"PlayLowest", ai.PlayLowest},
		NamedPlay{"PlayRandom", ai.PlayRandom},
	}
	players := make([]ai.Player, 0)
	for _, b := range bidding_strategies {
		for _, p := range playing_strategies {
			players = append(players, ai.Player{fmt.Sprintf("%v:%v", b.Name, p.Name), b.Bid, p.Play})
		}
	}
	return players
}

func createMatchRunners() {
	for x := 0; x < numberOfMatchRunners; x++ {
		go simulateMatches(matchesToSimulate / numberOfMatchRunners)
	}
}

func simulateMatches(matchesToSimulate int) (int, int) {
	for {
		oponents := <-oponents
		win1 := 0
		win2 := 0
		for x := 0; x < matchesToSimulate; x++ {
			winningPartnership, match := playMatch(oponents.player1, oponents.player2)
			if winningPartnership == 0 {
				win1++
			} else {
				win2++
			}
			if debugLog && x%10 == 0 {
				fmt.Printf("Current standings: %v - %v\n", win1, win2)
			}
			if debugLog {
				fmt.Printf("Partnership %v won! %v\n", winningPartnership, match)
			}
		}
		results <- Result{win1, win2}
	}
}

func playMatch(player1, player2 ai.Player) (int, *ai.Match) {
	var players [4]ai.Player
	for x := 0; x < 4; x++ {
		if x%2 == 0 {
			players[x] = player1
		} else {
			players[x] = player2
		}
	}
	match := ai.Match{
		Partnerships: new([2]ai.Partnership),
		Players:      players,
	}
	winner := -1
	for x := 1; winner == -1; x++ {
		bidder := playDeal(&match, x%4)
		pOne := match.Partnerships[0]
		pTwo := match.Partnerships[1]
		if pOne.MatchScore >= winningScore && pTwo.MatchScore >= winningScore {
			winner = bidder % 2
		} else if pOne.MatchScore >= winningScore || pTwo.MatchScore <= giveUpScore {
			winner = 0
		} else if pTwo.MatchScore >= winningScore || pTwo.MatchScore <= giveUpScore {
			winner = 1
		}
	}
	return winner, &match
}

func playDeal(match *ai.Match, dealer int) int {
	deck := sdz.CreateDeck()
	deck.Shuffle()
	hands := deck.Deal()

	match.Partnerships[0].DealScore = 0
	match.Partnerships[1].DealScore = 0

	for x, hand := range hands {
		match.Hands[x] = hand
		sort.Sort(hand)
		if debugLog {
			fmt.Println(hand)
			for _, suit := range sdz.Suits {
				meld, _ := hand.Meld(suit)
				fmt.Printf("%v: %v ", suit, meld)
			}
		}
	}
	bid, playerWithBid, trump := bid(match, dealer)
	playerWithLead := playerWithBid
	match.Partnerships[playerWithLead%2].Bid = bid
	match.Partnerships[(playerWithLead+1)%2].Bid = 0
	match.SetMeld(trump)
	if debugLog {
		fmt.Println("Trump:", trump)
		fmt.Println("bids", match.Partnerships[0].Bid, match.Partnerships[1].Bid)
		fmt.Println("deal scores", match.Partnerships[0].DealScore, match.Partnerships[1].DealScore)
	}
	for x := 0; x < numberOfTricks; x++ {
		playerWithLead, trick := playHand(match, playerWithLead, trump)
		if debugLog {
			fmt.Println("lead", playerWithLead, trick, "counters:", trick.Counters())
			fmt.Println("deal scores before:", match.Partnerships[0].DealScore, match.Partnerships[1].DealScore)
		}
		match.Partnerships[playerWithLead%2].DealScore += trick.Counters()
		match.Partnerships[playerWithLead%2].HasTakenTrick = true
		// Last trick
		if x+1 == numberOfTricks {
			match.Partnerships[playerWithLead%2].DealScore++
		}
		if debugLog {
			fmt.Println("deal scores after:", match.Partnerships[0].DealScore, match.Partnerships[1].DealScore)
		}
	}
	match.Partnerships[0].SetDealScore()
	match.Partnerships[1].SetDealScore()
	if debugLog {
		fmt.Println("Match after: ", match)
	}

	return playerWithBid
}

func playHand(match *ai.Match, playerWithLead int, trump sdz.Suit) (int, sdz.Hand) {
	winningCard := sdz.NACard
	winningPlayer := playerWithLead
	leadSuit := sdz.NASuit
	trick := make([]sdz.Card, 0)

	for x := playerWithLead; x < playerWithLead+4; x++ {
		currPlayer := x % 4
		currHand := &match.Hands[currPlayer]
		if debugLog {
			fmt.Println(currHand)
		}
		card := match.Players[currPlayer].Play(currHand, winningCard, leadSuit, trump)
		currHand.Remove(card)
		trick = append(trick, card)
		if winningCard == sdz.NACard {
			winningCard = card
			leadSuit = card.Suit()
		} else if card.Beats(winningCard, trump) {
			winningCard = card
			winningPlayer = currPlayer
		}
	}
	if debugLog {
		fmt.Println("Trick: ", trick)
		fmt.Println("Match: ", match)
	}
	return winningPlayer, trick
}

func bid(match *ai.Match, dealer int) (int, int, sdz.Suit) {
	var highBid int = 20
	var highBidder int = dealer
	var trump sdz.Suit
	bids := make([]int, 0)
	_, trump = match.Players[dealer].Bid(&match.Hands[dealer], bids)
	for bidder := dealer + 1; bidder < dealer+5; bidder++ {
		index := bidder % 4
		bid, suit := match.Players[index].Bid(&match.Hands[index], bids)
		if highBid == 20 && dealer == index {
			trump = suit
		} else if bid > highBid {
			highBid, trump = bid, suit
		}
		bids = append(bids, bid)
	}
	return highBid, highBidder, trump
}