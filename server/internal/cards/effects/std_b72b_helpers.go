package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b72b_helpers.go — shared pieces for Lorwyn Eclipsed batch 72b
// (#4458).

// KindredCommandCopyMode is the first bullet every Lorwyn Eclipsed
// kindred Command prints — "Create a token that's a copy of target
// <Subtype> you control." (Ashling's, Brigid's, Grub's and Trystan's
// Commands). The target is any permanent of that subtype the caster
// controls, a kindred noncreature one or a changeling included, read
// through the layered view; the token copies its copiable values
// (CR 707.2) and is created under the caster's control. Pass it to
// ChooseN with the card's other bullets.
func KindredCommandCopyMode(subtype string) game.ModeOption {
	return ModeDoing("Create a token that's a copy of target "+subtype+" you control.",
		TargetPermanent("target "+subtype+" you control", Subtype(subtype), YouControl()),
		TokenCopyTheModesTarget)
}

// TargetPlayerDrawsTwoMode is the bullet "Target player draws two
// cards." (Ashling's Command), declared as a draw for the bots.
func TargetPlayerDrawsTwoMode() game.ModeOption {
	return ModeWithPurpose(ModeDoing("Target player draws two cards.", TargetPlayer("target player"),
		TheModesTargetPlayerDrawsTwo),
		ForTargets(game.TargetPurpose{Slot: 0, Draws: 2}))
}

// TheModesTargetPlayerDrawsTwo is "target player draws two cards" as a
// modal bullet's body (Ashling's Command, Metamorphic Blast's spree
// bullet).
func TheModesTargetPlayerDrawsTwo(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return DrawCards{Player: t.ID, N: 2}.Apply(ctx)
}

// mutavaultOracleID is the card Mutavault's oracle ID (mutavault.go).
const mutavaultOracleID = "6b3cc59a-7ea5-4eb5-9bf9-5a9c07f80e2b"

// MutavaultToken is "a Mutavault token": a token created by name, with
// no other characteristics given, takes the characteristics of the card
// with that name in the Oracle card reference (CR 111.11; Mutable
// Explorer's ruling says the same). So the token carries Mutavault's
// oracle ID, exactly as a token copy carries the copied card's, and its
// "{T}: Add {C}" and "{1}: becomes a 2/2 with all creature types" are the
// catalog card's own rows. A colourless Token Land, as printed (tecl 11).
func MutavaultToken() game.Card {
	return game.Card{
		Name:         "Mutavault",
		OracleID:     mutavaultOracleID,
		TypeLine:     "Token Land",
		ProducedMana: []string{"C"},
	}
}
