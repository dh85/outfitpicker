import ArgumentParser

@main
struct OutfitPicker: ParsableCommand {
    static let configuration = CommandConfiguration(
        abstract: "Interactive CLI to pick outfits from category folders with per-category rotation"
    )
    
    func run() throws {
        print("OutfitPicker - Coming soon!")
    }
}