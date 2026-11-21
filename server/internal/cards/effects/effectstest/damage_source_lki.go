package effectstest

// damage_source_lki_test.go is the card-level proof for #1396: damage
// from a permanent that has LEFT the battlefield carries the lifelink
// and deathtouch it had as it last existed (CR 608.2h), and the object a
// trigger named is never confused with a new object the same card has
// become (CR 400.7). The #1379 half — the last-known POWER — is
// resolution_lki_test.go.

const ManaVaultOracle = "736892cb-a34b-4bb9-b56c-e26e3db207a2"
