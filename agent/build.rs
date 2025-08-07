fn main() {
    tonic_build::configure()
        .build_server(false)
        .compile(
            &["../backend/api/proto/guardian.proto"],
            &["../backend/api/proto"]
        )
        .expect("Failed to compile proto");
}
