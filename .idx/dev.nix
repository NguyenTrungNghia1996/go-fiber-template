{ pkgs, ... }: {
  channel = "stable-24.05";

  packages = [
    pkgs.go
    pkgs.air
    pkgs.ngrok
  ];

  env = {
    MONGO_URL = "mongodb+srv://admin:cr969bp6x6@test.nlwvliy.mongodb.net/?retryWrites=true&w=majority&appName=test";
    MONGO_NAME = "test2";
    JWT_SECRET = "test";
    PORT = "4000";
    MINIO_ACCESS_KEY = "al8KsxHAbLtfNVsX";
    MINIO_SECRET_KEY = "noWZ40KlvEcioZcPhLmMZFcPSkdeuX0K";
    MINIO_ENDPOINT = "image.nghia.myds.me";
    MINIO_BUCKET = "test";
    MINIO_SSL = "true";
    NGROK_AUTH_TOKEN = "1n391cGDskZaeIjGacWKh0NS93J_83n7ttgsiQJzVtV2R5nn1";
  };

  idx = {
    extensions = [
      "golang.go"
    ];

    previews = {
      enable = true;
      previews = {
        web = {
          command = ["air"];
          manager = "web";
          env = {
            PORT = "$PORT";
          };
        };
      };
    };

    workspace = {
      onCreate = {
        init = ''
          go mod tidy
        '';
      };
      onStart = {
        setup-ngrok = "ngrok config add-authtoken $NGROK_AUTH_TOKEN";
        start = "air";
        tunnel = "ngrok http 4000";
      };
    };
  };
}
