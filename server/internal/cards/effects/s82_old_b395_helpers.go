package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// OldB395ThisOrEnchantedDealtCombatDamageToAPlayer is the condition of
// "Whenever this creature or enchanted creature deals combat damage to a
// player" (Thassa's Emissary): the damage is combat damage to a player, dealt
// by the source or by the creature it is attached to.
func OldB395ThisOrEnchantedDealtCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if ev.Source != source.InstanceID && !source.IsAttachedTo(ev.Source) {
		return false
	}
	return g.PlayerByIDForEffect(ev.Target) != nil
}

// OldB395GainsUntilEOT is "This creature gains <keyword> until end of turn":
// nothing happens if the source is no longer this permanent (CR 400.7).
func OldB395GainsUntilEOT(label, keyword string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{keyword}, Label: label}.Apply(NewContext(g, item))
	}
}

// OldB395TapThis is "tap this creature" as a trigger body: the source taps if it
// is still on the battlefield as the trigger resolves.
func OldB395TapThis(g *game.Game, item *game.StackItem) error {
	if !B09SourceStillOnBattlefield(g, item) {
		return nil
	}
	return TapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
}

// OldB395GrantTargetsUntilEOT is "each of the targets gains <keywords> until
// end of turn": every target still legal as the ability resolves (CR 608.2b).
func OldB395GrantTargetsUntilEOT(name string, keywords ...string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: keywords, Label: name + " — gains keywords until end of turn"}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// OldB395MustBeBlockedThis is "This creature must be blocked this turn if
// able": nothing happens if the source is no longer this permanent (CR 400.7).
func OldB395MustBeBlockedThis(g *game.Game, item *game.StackItem) error {
	if !SourceIsStillThisPermanent(g, item) {
		return nil
	}
	return BlockRequirementUntilEOT{
		Target: item.SourceCardID,
		Kind:   game.BlockRequirementMustBeBlocked,
		Label:  "Loathsome Catoblepas — must be blocked this turn",
	}.Apply(NewContext(g, item))
}
