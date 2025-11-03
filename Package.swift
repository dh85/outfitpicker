// swift-tools-version: 6.1
import PackageDescription

let package = Package(
    name: "outfitpicker",
    platforms: [.macOS(.v13)],
    products: [
        .executable(name: "outfitpicker", targets: ["outfitpicker"])
    ],
    dependencies: [
        .package(url: "https://github.com/apple/swift-argument-parser", from: "1.3.0")
    ],
    targets: [
        .executableTarget(
            name: "outfitpicker",
            dependencies: [
                .product(name: "ArgumentParser", package: "swift-argument-parser")
            ]
        ),
        .testTarget(
            name: "outfitpickerTests",
            dependencies: ["outfitpicker"]
        )
    ]
)
