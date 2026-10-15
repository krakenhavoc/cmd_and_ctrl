package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Warring Triad — Legendary Artifact Creature — God {3}, 5/5
// (EDHREC rank 3283):
//
//	"Flying, trample, haste
//	 As long as there are fewer than eight cards in your graveyard,
//	 The Warring Triad isn't a creature.
//	 {T}, Mill a card: Target player adds one mana of any color.
//	 (Activate only as an instant.)"
//
// A three-mana 5/5 flying trampler with haste that is a plain
// artifact until the graveyard fills. The keywords ride
// PrintedKeywords. "Isn't a creature" is a layer 4 static on itself
// gated on the controller's graveyard (b31NotACreatureWhileGraveyardBelow):
// while the graveyard holds fewer than eight cards its Creature type
// is dropped — and its creature type God with it (CR 205.1b) — so it
// cannot attack or block, does not die to a wrath or to damage, and
// is not summoning sick when it wakes because it has been on the
// battlefield all along; Legendary and Artifact stay. Eight cards
// and it is a creature at once.
//
// The mana ability. "{T}, Mill a card" is MillCardsCost(1) beside the
// tap (#3845, ADR 0020's 2026-10-15 amendment), paid at announce
// (CR 602.2b, CR 601.2h): the top card goes to the graveyard through
// the ordinary mill, so Rest in Peace exiles it instead and "whenever
// you mill" payoffs see it, and a library of no cards can't pay it
// (CR 701.17b), so the ability is not offered then.
//
// It TARGETS a player, so it is NOT a mana ability (CR 605.1a): it
// uses the stack, an opponent may respond, and it can be countered.
// The printed "(Activate only as an instant.)" is the reminder that
// the ability is an ordinary one with no sorcery-speed restriction,
// and adds no rule of its own. The {T} is a real cost of a permanent
// that is not a creature while the graveyard is short, so summoning
// sickness (CR 302.6) does not apply to it then; once the graveyard
// holds eight cards and it is a creature, a Triad that came in this
// turn is sick and cannot pay it, which the engine enforces from the
// permanent's effective types.
//
// The TARGET player adds the mana (AddMana{Player: target}) and makes
// the colour choice, because the player who adds the mana is the one
// who chooses it: the pick is queued for that seat, and the mana lands
// in that seat's pool, so aiming it at an opponent hands them the pick.
// Handing an opponent mana is a real Commander play, so the card is
// written the way it is printed. (Radiant Lotus differs on the colour:
// it says "choose a color", which is the controller's.)
//
// Layer cache. The static above reads the graveyard, and a card
// reaching or leaving a graveyard from any zone but the battlefield
// bumps the layer version unconditionally (layer_listener.go, #1117),
// so the eighth card — milled by this very ability, discarded or
// resolved from the stack — wakes the Triad at once. The ability's
// own tap comes BEFORE its mill in the payment order, so what the tap
// recompute saw is the seven-card graveyard and it is the mill's
// arrival that wakes it. TestWarringTriadWakesOnTheEighthCard pins it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "df0b5995-117b-4ba8-964d-ea3c592621ef",
		Name:            "The Warring Triad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "haste"},
		Static: []game.StaticAbility{
			b31NotACreatureWhileGraveyardBelow(8),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Mill a card: Target player adds one mana of any color.",
			Cost:    Plus(TapCost(), MillCardsCost(1)),
			Targets: TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return AddMana{Player: t.ID, Produced: "{W|U|B|R|G}"}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
