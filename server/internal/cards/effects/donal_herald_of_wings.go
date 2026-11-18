package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Donal, Herald of Wings — Legendary Creature — Human Wizard {2}{U}{U}, 3/3
// (Innistrad: Crimson Vow):
//
//	"Whenever you cast a nonlegendary creature spell with flying, you may
//	 copy it, except the copy is a 1/1 Spirit in addition to its other
//	 types. Do this only once each turn. (The copy becomes a token.)"
//
// Iron Man, Bleeding Edge's shape: the "you may" is asked as the trigger
// resolves (CR 603.5) and "only once each turn" limits the DOING, so
// declining does not use the turn up. The copy is Tawnos's (CopySpell from
// last-known information, CR 707.10) whose exception sets the copiable
// power and toughness to 1/1 and adds Spirit (CR 707.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c9f2b6e-a0eb-493e-8c43-3a14f6d95a53",
		Name:         "Donal, Herald of Wings",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || b28DoneThisTurn(g, source.InstanceID, donalHeraldOfWingsDone) {
					return false
				}
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && And(Creature(), Not(Legendary()), HasKeyword("flying"))(g, source.Controller, game.ThroughTextOf(source, spell))
			}, donalHeraldOfWingsLabel, donalHeraldOfWingsCopy),
				"Donal, Herald of Wings — copy that spell?"),
		},
	})
}

const (
	donalHeraldOfWingsLabel = "Donal, Herald of Wings — copy that spell, except the copy is a 1/1 Spirit in addition to its other types"
	donalHeraldOfWingsDone  = "Donal, Herald of Wings — copied a spell this turn"
)

func donalHeraldOfWingsCopy(g *game.Game, item *game.StackItem) error {
	if b28DoneThisTurn(g, item.SourceCardID, donalHeraldOfWingsDone) {
		return nil
	}
	ctx := NewContext(g, item)
	b28NoteDoneThisTurn(g, item.SourceCardID, donalHeraldOfWingsDone)
	return CopySpell{
		StackID:       ctx.Trigger().Event.CardID,
		FromLastKnown: true,
		Except: func(v *game.PrintedValues) {
			v.SetPowerToughness(1, 1)
			v.AddSubtype("Spirit")
		},
	}.Apply(ctx)
}
