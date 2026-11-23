package effectstest

// rooms_seam_wave_test.go — the eight Rooms that waited on single-door
// mechanics (home tracker #2555). Each test drives a door that works;
// the doors declared unimplemented in their caveats have no behaviour
// to test, and TestRoomsSeamWaveDeclareTheirGaps pins the caveat so the
// gap cannot silently close behind a stale claim.

const (
	MoldGymOracle      = "34d0d070-fcaf-410e-a002-012cbe104fb7"
	TicketBoothOracle  = "4d01b62b-b924-4da5-8ff5-b2f29d7f19b2"
	UnderwaterOracle   = "2b46394e-d337-4b1a-88e6-7fdac2ee4ac4"
	ExpLabOracle       = "8869df0c-fe84-4964-8d38-a8ecefa9c252"
	CharredFoyerOracle = "9f50ec9f-78e6-403f-8f41-84d603fee0eb"
	CrampedVentsOracle = "1e847d69-e527-4bb9-a624-17c25ac4fed4"
	DazzlingOracle     = "c46a02db-13d6-477f-9da0-822599470168"
	SecretArcadeOracle = "c8abde48-a07a-42a0-a43b-c357e6d9cad4"
)
