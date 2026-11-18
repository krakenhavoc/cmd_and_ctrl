package game

import "github.com/google/uuid"

// autotap_zone_costed.go — #5896: the costed auto-tap pass (#2455)
// for a mana ability that functions OFF the battlefield (CR 113.6b).
//
//	Jack-o'-Lantern  "{1}, Exile this card from your graveyard: Add one
//	                  mana of any color."
//
// The ordinary off-battlefield picker (autoManaExileAbilityFor) takes
// only an ability whose whole cost is the exile, because it has no way
// to fund mana. The costed solver does (autotap_costed.go): it orders
// the sources so a costed one's mana is paid out of mana made before
// it, never out of its own. An off-battlefield source joins that chain
// exactly as a Signet does; what differs is the payment the executor
// makes after the mana — the card leaves its zone instead of tapping —
// which materializeExiledManaSourceLocked already makes for the
// Spirit Guides.

// autoManaExileCostedFor is autoManaExileAbilityFor for an ability that
// owes MANA besides its exile: every demand that picker makes but the
// mana one (autoManaExileAccepts), and the mana priced by the same rule
// the battlefield costed picker uses (autoTapCostedPrice). Returns the
// ability and its priced cost, or nil.
//
// Shared by the planner (gatherManaZoneSources with costed set) and the
// executor (materializeExiledManaSourceLocked for a Costed entry), so
// the two cannot disagree about which ability a card is spent for.
//
// Caller must hold g.mu.
func (g *Game) autoManaExileCostedFor(asker uuid.UUID, source Card, abilities []ManaAbilityShape, zone ZoneKind) (*ManaAbilityShape, ParsedCost) {
	if zone == ZoneBattlefield {
		return nil, ParsedCost{}
	}
	for i := range abilities {
		a := abilities[i]
		if !g.autoManaExileAccepts(asker, source, a, zone) {
			continue
		}
		if !g.ManaAbilityChargesManaForEffect(asker, source, a) {
			continue
		}
		cost, ok := g.autoTapCostedPrice(asker, source, a)
		if !ok {
			continue
		}
		return &a, cost
	}
	return nil, ParsedCost{}
}

// zoneManaAbilityForPlanLocked is the ability a plan entry for an
// off-battlefield source booked: the costed picker for a Costed entry,
// the ordinary one for every other. The executor and the preview both
// re-find it through here.
//
// Caller must hold g.mu.
func (g *Game) zoneManaAbilityForPlanLocked(asker uuid.UUID, source Card, planned plannedTap, zone ZoneKind) *ManaAbilityShape {
	if planned.Costed {
		ab, _ := g.autoManaExileCostedFor(asker, source, ManaAbilitiesForCard(source), zone)
		return ab
	}
	return g.autoManaExileAbilityFor(asker, source, ManaAbilitiesForCard(source), zone)
}
