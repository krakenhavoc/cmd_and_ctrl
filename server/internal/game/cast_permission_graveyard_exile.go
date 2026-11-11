package game

// AltCostKeyGraveyardExile is the offer a standing graveyard cast
// permission prices a cast under when the permission asks for cards to be
// exiled from the graveyard IN ADDITION to the spell's own costs (Kotis,
// Sibsig Champion: "by exiling three other cards from your graveyard in
// addition to paying its other costs"). It is not escape: the spell keeps
// its printed mana cost, and it is never "escaped" (CR 702.138). The
// number of cards is CastPermission.ExileOtherFromGraveyard.
const AltCostKeyGraveyardExile = "graveyard-exile"
