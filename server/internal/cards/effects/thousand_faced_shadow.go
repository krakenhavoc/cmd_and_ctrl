package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thousand-Faced Shadow — {U} Creature — Human Ninja 1/1 (Kamigawa: Neon
// Dynasty, S82 NEO batch 20):
//
//	"Ninjutsu {2}{U}{U}
//	 Flying
//	 When this creature enters from your hand, if it's attacking, create a
//	 token that's a copy of another target attacking creature. The token
//	 enters tapped and attacking."
//
// Both halves of the "if" are about the entry: it came from a hand (the
// entry's zone change names the hand a ninjutsu puts it onto the battlefield
// from, CR 603.6a; a spell cast from the hand enters from the stack) and it
// is attacking. "It's attacking" is checked as the creature enters
// and again as the ability resolves (CR 603.4); where it came from is
// history and is not asked again. The copy is made from the target's
// copiable values (CR 707.2) and enters tapped and attacking: its
// controller chooses the player or planeswalker it attacks, the copied
// creature's own defender first (CR 508.4). A Shadow that is not attacking
// when it enters (hard-cast) triggers nothing and asks for no target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "98636623-bbc6-4935-a2bb-2234691eb6d8",
		Name:            "Thousand-Faced Shadow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{Ninjutsu("{2}{U}{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(InterveningIf(thousandFacedShadowAttacking,
				On(game.EventETB, AllOf(Self, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return thousandFacedShadowEnteredFromHand(g, source.InstanceID) && source.AttackingTarget != uuid.Nil
				}),
					"Thousand-Faced Shadow — create a token that's a copy of another target attacking creature, tapped and attacking",
					Do(Instruction(thousandFacedShadowCopy)))),
				Another(TargetCreature("another target attacking creature", AttackingCreature()))),
		},
	})
}

var thousandFacedShadowAttacking = CondBoard{
	Name: "Thousand-Faced Shadow is attacking",
	Test: func(ctx *Context) bool {
		c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
		return ok && onBattlefield(ctx.Game, ctx.Source()) && c.AttackingTarget != uuid.Nil
	},
}

func thousandFacedShadowCopy(ctx *Context) error {
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	tmpl, ok := TokenCopyTemplate(ctx.Game, id)
	if !ok {
		return nil
	}
	prefer := uuid.Nil
	if c, ok := ctx.Game.LookupCardForEffect(id); ok {
		prefer = c.AttackingTarget
	}
	return CreateTokensAttackingYourChoice{
		Template: tmpl,
		Copy:     true,
		N:        1,
		Tapped:   true,
		Prefer:   prefer,
		Label:    "Thousand-Faced Shadow",
	}.Apply(ctx)
}

// thousandFacedShadowEnteredFromHand is "enters from your hand": the card's
// latest move onto the battlefield left a hand, not the stack or another
// zone.
func thousandFacedShadowEnteredFromHand(g *game.Game, id uuid.UUID) bool {
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := g.Events[i]
		if ev.Kind == game.EventZoneMove && ev.CardID == id && ev.NewZone == game.ZoneBattlefield {
			return ev.OldZone == game.ZoneHand
		}
	}
	return false
}
