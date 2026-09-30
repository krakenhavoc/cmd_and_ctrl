package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Myr Battlesphere — Artifact Creature — Myr Construct {7}:
//
//	"When this creature enters, create four 1/1 colorless Myr artifact
//	 creature tokens.
//	 Whenever this creature attacks, you may tap X untapped Myr you
//	 control. If you do, this creature gets +X/+0 until end of turn and
//	 deals X damage to the player or planeswalker it's attacking."
//
// "You may tap X untapped Myr" is a choice of any number of them
// (including none), made as the trigger resolves. Every pick is
// re-checked before it is tapped, because the board moves under an
// asynchronous prompt, and X is how many actually became tapped —
// which is also why summoning-sick Myr tokens are fair game: tapping
// them is an effect, not a {T} cost. The Battlesphere itself can be
// one of them only if it is untapped (vigilance).
//
// The damage goes to the player or planeswalker this attack is aimed
// at, read off the attack event; an attack on a battle deals no damage
// (only the +X/+0 applies), which is the printed text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c53ba31a-ba27-4e17-9a92-311acb1cab29",
		Name:         "Myr Battlesphere",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Myr Battlesphere — create four 1/1 colorless Myr artifact creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 colorless Myr artifact"), N: 4})),
			WheneverThisAttacks("Myr Battlesphere — you may tap X untapped Myr you control", myrBattlesphereAttack),
		},
	})
}

// myrBattlesphereAttack offers the controller's untapped Myr.
func myrBattlesphereAttack(g *game.Game, item *game.StackItem) error {
	var myr []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && !c.Tapped && c.HasSubtype("Myr") {
			myr = append(myr, c.InstanceID)
		}
	}
	if len(myr) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Myr Battlesphere — tap X untapped Myr you control",
		Cards:    myr,
		Min:      0,
		Max:      0, // "any number"
		Zone:     game.ZoneBattlefield,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			return myrBattlesphereTapAndHit(g, item, picked)
		},
	})
	return nil
}

// myrBattlesphereTapAndHit taps the picks that are still untapped Myr
// and, if any were, pumps the Battlesphere and hits what it attacks.
func myrBattlesphereTapAndHit(g *game.Game, item *game.StackItem, picked []uuid.UUID) error {
	ctx := NewContext(g, item)
	x := 0
	for _, id := range picked {
		c, ok := g.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != item.Controller || !c.HasSubtype("Myr") {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		x++
	}
	if x == 0 {
		return nil
	}
	if err := (BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: "Myr Battlesphere — +X/+0"}).Apply(ctx); err != nil {
		return err
	}
	attacked := item.Trigger.Event.Target
	if g.PlayerByIDForEffect(attacked) == nil {
		if c, ok := g.LookupCardForEffect(attacked); !ok || !c.IsPlaneswalker() {
			return nil
		}
	}
	return DealDamage{Source: item.SourceCardID, Target: attacked, Amount: x}.Apply(ctx)
}
