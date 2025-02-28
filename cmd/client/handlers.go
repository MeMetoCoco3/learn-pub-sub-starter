package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerLogs() func(routing.GameLog) pubsub.AckType {
	return func(logg routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		gamelogic.WriteLog(logg)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, chann *amqp091.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Println("> ")
		moveOutcome := gs.HandleMove(move)

		switch moveOutcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJson(chann,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				log.Printf("Error: %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("Error: Unknown Move outcome")
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, chann *amqp091.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(dw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		warOutcome, winner, loser := gs.HandleWar(dw)
		msg := ""
		switch warOutcome {
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
			msg = winner + " won a war against " + loser
		case gamelogic.WarOutcomeDraw:
			msg = "A war between " + winner + " and " + loser + " resulted in a draw"
		}

		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon, gamelogic.WarOutcomeDraw:

			// Sending Battle Logs
			gamelog := routing.GameLog{
				CurrentTime: time.Now(), Message: msg, Username: gs.Player.Username,
			}
			log.Println(dw.Attacker.Username)
			log.Println(routing.GameLogSlug)
			log.Println(dw.Attacker.Username)
			if err := pubsub.PublishGob(chann,
				routing.ExchangePerilTopic,
				routing.GameLogSlug+"."+gs.Player.Username,
				gamelog,
			); err != nil {
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}

		log.Println("Uknown war outcome")
		return pubsub.NackDiscard
	}
}
