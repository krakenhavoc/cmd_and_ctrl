package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invasion of Innistrad — Battle — Siege, defense 5, for {2}{B}{B}:
//
//	"Flash
//	 When this Siege enters, target creature an opponent controls
//	 gets -13/-13 until end of turn.
//	 (As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)"
//
// The reminder text in brackets is not card-effect data: the
// protector prompt and the defense counters are engine behaviour
// keyed on the card TYPE (server/internal/game/battle.go), so they
// would happen to this card even with no Spec at all. What the Spec
// carries is the flash, the ETB removal, and the defeated trigger.
//
// The defeated trigger exiles the Siege and grants its controller a
// free cast of Deluge of the Dead, which lives in its own file under
// the "#1" catalog key. S27 shipped this card with the exile alone
// and the transformed cast declared missing; S32's per-instance face
// on CastPermission closed that. What remains is the timing —
// see SiegeTransformedCastCaveat and SiegeDefeated in battles.go.
//
// -13/-13 rather than "destroy": the difference is observable and it
// is the reason the card is played. A creature with indestructible
// dies to this, and one that regenerates does not come back — it is
// not destruction, it is a toughness reduction the CR 704.5f
// state-based action answers.
// invasionOfInnistradOracleID is shared with the back face's spec,
// which registers under this ID plus "#1" — the one place the two
// halves of a double-faced card have to agree on a string.
const invasionOfInnistradOracleID = "a3c1af66-63c8-41ec-a401-a3da1131dc67"

func init() {
	Register(Spec{
		OracleID:        invasionOfInnistradOracleID,
		Name:            "Invasion of Innistrad",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{SiegeTransformedCastCaveat},
		PrintedKeywords: []string{"flash"},
		Battle: &BattleSpec{
			Defense: 5,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.CardID == source.InstanceID
				},
				Targets: TargetCreature("target creature an opponent controls", OpponentControls()),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Invasion of Innistrad — -13/-13",
						func(g *game.Game, item *game.StackItem) error {
							if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
								return nil
							}
							return BoostUntilEOT{
								Target:    item.Targets[0].ID,
								Power:     -13,
								Toughness: -13,
								Label:     "Invasion of Innistrad — -13/-13",
							}.Apply(NewContext(g, item))
						})
				},
			},
			DefeatedTrigger("Invasion of Innistrad — defeated: exile it", SiegeDefeated()),
		},
	})
}
