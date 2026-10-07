package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gideon_animate.go — the shared body of the Gideon planeswalkers'
// "Until end of turn, Gideon becomes a P/T <types> creature [with
// indestructible] that's still a planeswalker. Prevent all damage that
// would be dealt to him this turn." (#2046, ADR 0032 amendment of
// 2026-10-07.)
//
// Two data records, both about the object that activated:
//
//   - THE BECOMING is one ScopedEffectFor pinned to Gideon until end of
//     turn: layer 4 adds the Creature type and the subtypes, layer 7b
//     sets the base power and toughness, layer 6 adds indestructible
//     when the card prints it. He keeps his planeswalker type (nothing
//     removes it), so damage to him has both results (CR 120.3c and
//     120.3e) and both state-based actions apply (CR 704.5g and 704.5i).
//   - THE SHIELD is ADR 0108 §7's source shield with no source,
//     protecting Gideon himself this turn. It is a prevention effect,
//     so CR 615.12 lets damage that can't be prevented through it, which
//     is the case the double type exists for.
//
// A Gideon that left the battlefield in response is a new object
// (CR 400.7) and the ability does nothing.

// gideonAnimation is what one Gideon's becoming says.
type gideonAnimation struct {
	// Label names the ability in the log.
	Label string
	// Subtypes are the creature types he gains ("Human", "Soldier").
	Subtypes []string
	// Colors, when non-empty, is "a white Soldier creature" (layer 5).
	Colors []string
	// Power and Toughness are the base P/T the becoming sets (layer 7b).
	Power, Toughness int
	// Indestructible is "with indestructible".
	Indestructible bool
}

// animateGideon applies the becoming to the ability's source and then
// the shield that protects him this turn.
func animateGideon(g *game.Game, item *game.StackItem, a gideonAnimation) error {
	ctx := NewContext(g, item)
	self := item.SourceCardID
	if z := g.FindCardZoneForEffect(self); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	mods := []game.Mod{
		game.AddTypesMod("Creature"),
		game.AddSubtypesMod(a.Subtypes...),
		game.SetBasePowerMod(a.Power),
		game.SetBaseToughnessMod(a.Toughness),
	}
	if len(a.Colors) > 0 {
		mods = append(mods, game.SetColorsMod(a.Colors...))
	}
	if a.Indestructible {
		mods = append(mods, game.AddKeywordsMod("indestructible"))
	}
	if err := (ScopedEffectFor{
		Target:   self,
		Mods:     mods,
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    a.Label,
	}).Apply(ctx); err != nil {
		return err
	}
	g.PreventDamageFromSourceThisTurnForEffect(game.DamageShield{
		EffectSource:     self,
		Controller:       item.Controller,
		ProtectPermanent: self,
		Label:            a.Label + " — prevent all damage that would be dealt to him this turn",
	})
	return nil
}
