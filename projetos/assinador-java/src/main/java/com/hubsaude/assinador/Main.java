package com.hubsaude.assinador;

import com.hubsaude.assinador.domain.SignRequest;
import com.hubsaude.assinador.domain.ValidateRequest;
import com.hubsaude.assinador.domain.SignatureResponse;

public class Main {

    private static final SignatureService service = new FakeSignatureService();

    public static void main(String[] args) {
        if (args.length == 0) {
            System.err.println("Erro: nenhum comando fornecido.");
            System.err.println("Uso: assinador <comando> [opções]");
            System.err.println("Comandos disponíveis: sign, validate");
            System.exit(1);
        }

        switch (args[0]) {
            case "sign"     -> handleSign(args);
            case "validate" -> handleValidate(args);
            default -> {
                System.err.println("Erro: comando desconhecido: " + args[0]);
                System.err.println("Comandos disponíveis: sign, validate");
                System.exit(1);
            }
        }
    }

    private static void handleSign(String[] args) {
        SignRequest request = new SignRequest();

        for (int i = 1; i < args.length; i++) {
            if ("--content".equals(args[i]) && i + 1 < args.length) {
                request.setContent(args[++i]);
            }
        }

        SignatureResponse response = service.sign(request);
        printResponse(response);
    }

    private static void handleValidate(String[] args) {
        ValidateRequest request = new ValidateRequest();

        for (int i = 1; i < args.length; i++) {
            switch (args[i]) {
                case "--content"   -> { if (i + 1 < args.length) request.setContent(args[++i]); }
                case "--signature" -> { if (i + 1 < args.length) request.setSignature(args[++i]); }
            }
        }

        SignatureResponse response = service.validate(request);
        printResponse(response);
    }

    private static void printResponse(SignatureResponse response) {
        String json = toJson(response);
        if (response.isValid()) {
            System.out.println(json);
        } else {
            System.err.println(json);
            System.exit(1);
        }
    }

    static String toJson(SignatureResponse r) {
        String sig = r.getSignature() == null
            ? "null"
            : "\"" + escapeJson(r.getSignature()) + "\"";
        return "{\"signature\":" + sig
            + ",\"valid\":" + r.isValid()
            + ",\"message\":\"" + escapeJson(r.getMessage()) + "\"}";
    }

    private static String escapeJson(String s) {
        return s.replace("\\", "\\\\").replace("\"", "\\\"");
    }
}
