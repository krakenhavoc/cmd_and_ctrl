package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roe_b1_helpers.go — shared shapes for Rise of the Eldrazi, batch 1 (#5581,
// S82).

// roeB1Spawns creates n of the catalog's Eldrazi Spawn tokens: 0/1
// colorless Eldrazi Spawn creatures with "Sacrifice this token: Add {C}."
func roeB1Spawns(n int) Applier {
	return CreateToken{Template: EldraziSpawnToken(), N: n}
}

// roeB1SpawnsWhenThisEnters is "When this creature enters, create n 0/1
// colorless Eldrazi Spawn creature tokens" (Emrakul's Hatcher, Kozilek's
// Predator, Nest Invader, Dread Drone).
func roeB1SpawnsWhenThisEnters(name string, n int) game.TriggeredAbility {
	label := name + " — create an Eldrazi Spawn"
	if n > 1 {
		label = name + " — create " + roeB1Count(n) + " Eldrazi Spawn"
	}
	return WhenThisEnters(label, Do(roeB1Spawns(n)))
}

// roeB1Count spells a small count as the cards print it.
func roeB1Count(n int) string {
	switch n {
	case 2:
		return "two"
	case 3:
		return "three"
	case 5:
		return "five"
	}
	return "some"
}

// roeB1YouControlAnEldraziSpawn is Brood Birthing's "if you control an
// Eldrazi Spawn": a permanent its controller controls that is both an
// Eldrazi and a Spawn, read as the spell resolves. Both words are read
// through the spell's text (CR 612.2).
func roeB1YouControlAnEldraziSpawn(ctx *Context) bool {
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.HasSubtype(ctx.TextSubtype("Eldrazi")) && c.HasSubtype(ctx.TextSubtype("Spawn")) {
			return true
		}
	}
	return false
}

// roeB1APlayerSacrificedACreature is "whenever a player sacrifices a
// creature": any player's sacrifice of a permanent that was a creature as it
// last stood (the event fires while it is still on the battlefield,
// CR 603.10a).
func roeB1APlayerSacrificedACreature(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventSacrifice || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature()
}

// roeB1CreaturesYouControl counts the creatures the source's controller
// controls (Pennon Blade's "for each creature you control").
func roeB1CreaturesYouControl(g *game.Game, source *game.Card) int {
	return permanentsYouControl(g, source, func(c *game.Card) bool { return c.IsCreature() })
}

// roeB1PreventCombatDamageToThis is "Prevent all combat damage that would be
// dealt to this creature" (CR 615.1a; Guard Gomazoa), Seraph of the Sword's
// replacement.
func roeB1PreventCombatDamageToThis(name string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:    []game.EventKind{game.EventDealDamage},
		Prevention: true,
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventDamage && ev.IsCombatDamage && ev.DamageTarget == src.InstanceID
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.Cancel()
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: name + " — prevent all combat damage that would be dealt to it",
	}
}

// roeB1ShrinkGrant is a leveler band's "{B}, {T}: Target creature gets
// -N/-N until end of turn." as a bundle (Guul Draz Assassin).
func roeB1ShrinkGrant(key, name string, n int) AbilityGrant {
	text := "{B}, {T}: Target creature gets " + signedPT(-n, -n) + " until end of turn."
	return AbilityGrant{
		Key: key,
		Activated: []ActivatedAbility{{
			Label:   text,
			Cost:    Plus(ManaCost("{B}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(HarmUntilEOT(0, -n, -n)),
			Effect:  pumpAndGrantFirstTarget(-n, -n, name+" — "+signedPT(-n, -n)+" until end of turn"),
		}},
		Text: text,
	}
}

// roeB1ElephantGrant is a leveler band's "{T}: Create N 3/3 green Elephant
// creature tokens." as a bundle (Kazandu Tuskcaller).
func roeB1ElephantGrant(key string, n int, text string) AbilityGrant {
	return AbilityGrant{
		Key: key,
		Activated: []ActivatedAbility{{
			Label:   text,
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: n, CreatureTokens: n},
			Cost:    TapCost(),
			Effect:  Do(CreateToken{Template: TokenCard("3/3 green Elephant"), N: n}),
		}},
		Text: text,
	}
}

// roeB1PingGrant is a leveler band's "{T}: This creature deals N damage to
// any target." as a bundle (Brimstone Mage). The host deals the damage.
func roeB1PingGrant(key string, n int) AbilityGrant {
	text := "{T}: This creature deals " + roeB1Digits(n) + " damage to any target."
	return AbilityGrant{
		Key: key,
		Activated: []ActivatedAbility{{
			Label:   text,
			Cost:    TapCost(),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, n)),
			Effect:  sourceDealsDamageToEachLegalTarget(n),
		}},
		Text: text,
	}
}

// roeB1Digits writes a small damage amount as a digit.
func roeB1Digits(n int) string {
	return string(rune('0' + n))
}
