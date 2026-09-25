package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// RMS Titanic — Legendary Artifact — Vehicle, {3}{R}:
//
//	"Flying, trample
//	 When RMS Titanic deals combat damage to a player, sacrifice it
//	 and create that many Treasure tokens.
//	 Crew 3"
//
// "That many" is the damage actually dealt — read off the event and
// captured by VALUE in Build, the same shape Old Gnawbone uses for
// its identical "create that many Treasures" clause. RMS Titanic
// adds one step ahead of the tokens: sacrificing itself, which is why
// it isn't just a call to Old Gnawbone's helper — the two bodies
// differ by more than a rename.
//
// Sacrificing first and then creating the tokens off the captured
// amount is safe even though the source is gone by the time the
// tokens appear: item.Controller is fixed at trigger creation and
// doesn't depend on the permanent still being on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6754753d-790e-4a95-a150-76741c4a02d4",
		Name:            "RMS Titanic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: ThisDealtCombatDamageToAPlayer,
			Key:       "RMS Titanic — sacrifice it and create that many Treasure tokens",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "RMS Titanic — sacrifice it and create that many Treasure tokens", nil)
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: rmsTitanicSacrificeAndTreasure,
		}},
		Activated: []ActivatedAbility{{
			Label:  "Crew 3",
			Cost:   CrewCost(3),
			Effect: CrewEffect("RMS Titanic"),
		}},
	})
}

// rmsTitanicSacrificeAndTreasure is the trigger's Effect: sacrifice the
// source, then create as many Treasure tokens as it dealt combat
// damage, stamped onto item.Params.Amount by a fill-in Build when the
// trigger fired.
func rmsTitanicSacrificeAndTreasure(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (SacrificePermanent{Target: item.SourceCardID}).Apply(ctx); err != nil {
		return err
	}
	return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: item.Params.Amount}.Apply(ctx)
}
