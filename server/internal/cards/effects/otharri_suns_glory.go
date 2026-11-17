package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Otharri, Suns' Glory — Legendary Creature — Phoenix {3}{R}{W}, 3/3 (March
// of the Machine: Phyrexia: All Will Be One Commander):
//
//	"Flying, lifelink, haste
//	 Whenever Otharri attacks, you get an experience counter. Then create a
//	 2/2 red Rebel creature token that's tapped and attacking for each
//	 experience counter you have.
//	 {2}{R}{W}, Tap an untapped Rebel you control: Return this card from your
//	 graveyard to the battlefield tapped."
//
// The experience counter is a player counter (CR 122.1), given before the
// tokens are counted, so the first attack makes one Rebel. The Rebels are
// created tapped and attacking (CR 508.4): the controller chooses which
// player or planeswalker each one attacks, the creature that triggered the
// ability first (CreateTokensAttackingYourChoice). They were never declared
// as attackers, so nothing that triggers on attacking sees them. The return
// ability works from the graveyard only; its cost taps an untapped Rebel
// they control, which is not the {T} symbol (CR 302.6) and so may be a Rebel
// that arrived this turn. Otharri is a Phoenix and cannot pay it herself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b081352b-d17b-4031-bbcc-c5c02cbfa6b6",
		Name:            "Otharri, Suns' Glory",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink", "haste"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Otharri, Suns' Glory — you get an experience counter, then create a tapped and attacking Rebel for each experience counter you have",
				func(g *game.Game, item *game.StackItem) error {
					if err := youGetAnExperienceCounter(g, item); err != nil {
						return err
					}
					n := experienceCounters(g, item.Controller)
					if n <= 0 {
						return nil
					}
					return CreateTokensAttackingYourChoice{
						Template: TokenCard("2/2 red Rebel"),
						N:        n,
						Tapped:   true,
						Prefer:   triggerAttackTarget(item),
						Label:    "Otharri, Suns' Glory",
					}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:  "{2}{R}{W}, Tap an untapped Rebel you control: Return this card from your graveyard to the battlefield tapped.",
			Cost:   Plus(ManaCost("{2}{R}{W}"), TapAnUntapped("an untapped Rebel you control", HasSubtype("Rebel"), YouControl())),
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: returnThisFromGraveyardTapped,
		}},
	})
}
