package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

func init() {
	cost := effects.TapAnUntapped("five untapped Zombies you control", effects.HasSubtype("Zombie"))
	cost.TapOthers.Count = 5
	effects.Register(effects.Spec{OracleID: "4235c99b-50de-4190-8938-3ce5ec5c5248", Name: "Gravespawn Sovereign", Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "Tap five untapped Zombies you control: Put target creature card from a graveyard onto the battlefield under your control.",
			Cost:  cost, Targets: effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
			Effect: effects.ReturnFirstLegalGraveyardTargetUnderYourControl,
		}},
	})
}
