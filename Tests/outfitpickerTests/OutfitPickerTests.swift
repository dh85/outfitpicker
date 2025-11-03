import Testing
@testable import outfitpicker

@Test("Basic functionality test")
func testBasicFunctionality() {
    // This is a placeholder test
    #expect(true == true)
}

@Test("Command configuration test") 
func testCommandConfiguration() {
    let config = OutfitPicker.configuration
    #expect(config.abstract.contains("Interactive CLI"))
}