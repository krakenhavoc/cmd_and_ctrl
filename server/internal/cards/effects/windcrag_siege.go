package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Windcrag Siege — Enchantment {1}{R}{W}:
//
//	"As this enchantment enters, choose Mardu or Jeskai.
//	 • Mardu — If a creature attacking causes a triggered ability of
//	   a permanent you control to trigger, that ability triggers an
//	   additional time.
//	 • Jeskai — At the beginning of your upkeep, create a 1/1 red
//	   Goblin creature token. It gains lifelink and haste until end
//	   of turn."
//
// #1647: the proof card for the gated TriggerDoubler. Mardu's line is
// word-for-word Isshin, Two Heavens as One's oracle text — the same
// DoublesAttacking(Creature()) doubler — but a TriggerDoubler is not
// an ability-list entry, so TriggersForCard / StaticAbilitiesForCard
// never see it and the ADR 0071 chosen-option gate that switches
// Frostcliff Siege's lines on and off cannot reach it that way. It
// gets its own ActiveWhen field instead (trigger_doubling.go), checked
// directly in triggerDoublersLocked against the doubler's own card —
// so a Jeskai Siege, or an unanswered one, doubles nothing.
//
// Jeskai is Legion Warboss's token one line over: the identical 1/1
// red Goblin, but lifelink and haste rather than haste and an attack
// requirement, so it is not required to attack and the ScopedEffectFor
// mods differ.
//
// No simplification.
func init() {
	mardu := DoublerWhenChosen("Mardu", DoublesAttacking(Creature()))
	mardu.Label = "Windcrag Siege"
	Register(Spec{
		OracleID:        "64df560a-905f-45d1-bc70-14d99e3112d3",
		Name:            "Windcrag Siege",
		Completeness:    CompletenessFull,
		AsEnters:        ChooseOptionAsEnters("Windcrag Siege", "Mardu", "Jeskai"),
		TriggerDoublers: []game.TriggerDoubler{mardu},
		Triggered: []game.TriggeredAbility{
			WhenChosen("Jeskai", AtYourUpkeep(
				"Windcrag Siege — create a 1/1 red Goblin with lifelink and haste until end of turn",
				windcragSiegeToken)),
		},
	})
}

// windcragSiegeToken creates the 1/1 red Goblin and grants lifelink
// and haste to that ONE token until end of turn (ScopedEffectFor, not
// a static — a permanent lord would pump every Goblin the controller
// has, not just this turn's).
func windcragSiegeToken(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cursor := b25LastEventSeq(g)
	if err := (CreateToken{Template: RedGoblinToken(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScopedEffectFor{
		Target:   tokens[0],
		Mods:     []game.Mod{game.AddKeywordsMod("lifelink", "haste")},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Windcrag Siege — the token has lifelink and haste",
	}.Apply(ctx)
}
