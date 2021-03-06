package claims

/*
// TMP commented until ClaimLinkObjectIdentity is updated to new spec
func TestClaimLinkObjectIdentity(t *testing.T) {
	// If generateTest is true, the checked values will be used to generate a test vector
	generateTest := false
	// Init test
	if err := testgen.InitTest("claimLinkObjectIdentity", generateTest); err != nil {
		panic(fmt.Errorf("error initializing test data: %w", err))
	}
	// Add input data to the test vector
	if generateTest {
		testgen.SetTestValue("idString", "113kyY52PSBr9oUqosmYkCavjjrQFuiuAw47FpZeUf")
		testgen.SetTestValue("objectHash", hex.EncodeToString([]byte{
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b,
			0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0c}))
		testgen.SetTestValue("auxData", hex.EncodeToString([]byte{
			0x0f, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
			0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x09,
			0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x01, 0x02}))
	}
	// ClaimLinkObjectIdentity
	const objectType = ObjectTypeAddress
	var indexType uint16
	id, err := core.IDFromString(testgen.GetTestValue("idString").(string))
	assert.Nil(t, err)
	var objectHash [256 / 8]byte
	var auxData [256 / 8]byte
	objectHashHex, _ := hex.DecodeString(testgen.GetTestValue("objectHash").(string))
	auxDataHex, _ := hex.DecodeString(testgen.GetTestValue("auxData").(string))
	copy(objectHash[:], objectHashHex[:256/8])
	copy(auxData[:], auxDataHex[:256/8])

	claim, err := NewClaimLinkObjectIdentity(objectType, indexType, id, objectHash, auxData)
	assert.Nil(t, err)
	claim.Version = 1
	e := claim.Entry()
	// Check claim against test vector
	checkClaim(e, t)
	dataTestOutput(&e.Data)
	c1 := NewClaimLinkObjectIdentityFromEntry(e)
	c2, err := NewClaimFromEntry(e)
	assert.Nil(t, err)
	assert.Equal(t, claim, c1)
	assert.Equal(t, claim, c2)
	// Stop test (write new test vector if needed)
	if err := testgen.StopTest(); err != nil {
		panic(fmt.Errorf("Error stopping test: %w", err))
	}
}
*/
