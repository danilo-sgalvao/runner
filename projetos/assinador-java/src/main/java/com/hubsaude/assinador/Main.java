package com.hubsaude.assinador;

public class Main {

    public static void main(String[] args) {
        if (args.length == 0) {
            System.err.println("Erro: nenhum comando fornecido.");
            System.err.println("Uso: assinador <comando> [opções]");
            System.err.println("Comandos disponíveis: sign, validate");
            System.exit(1);
        }

        String comando = args[0];

        try {
            switch (comando) {
                case "sign"     -> handleSign(args);
                case "validate" -> handleValidate(args);
                default -> {
                    System.err.println("Erro: comando desconhecido: " + comando);
                    System.err.println("Comandos disponíveis: sign, validate");
                    System.exit(1);
                }
            }
        } catch (IllegalArgumentException e) {
            System.err.println("Erro: " + e.getMessage());
            System.exit(1);
        }
    }

    private static void handleSign(String[] args) {
        String conteudo = null;
        String algoritmo = "SHA256withRSA";

        for (int i = 1; i < args.length; i++) {
            switch (args[i]) {
                case "--content"   -> { if (i + 1 < args.length) conteudo  = args[++i]; }
                case "--algorithm" -> { if (i + 1 < args.length) algoritmo = args[++i]; }
            }
        }

        String assinatura = AssinadorService.sign(conteudo, algoritmo);

        System.out.println("status=sucesso");
        System.out.println("assinatura=" + assinatura);
        System.out.println("algoritmo=" + algoritmo);
    }

    private static void handleValidate(String[] args) {
        String conteudo   = null;
        String assinatura = null;

        for (int i = 1; i < args.length; i++) {
            switch (args[i]) {
                case "--content"   -> { if (i + 1 < args.length) conteudo   = args[++i]; }
                case "--signature" -> { if (i + 1 < args.length) assinatura = args[++i]; }
            }
        }

        boolean valida = AssinadorService.validate(conteudo, assinatura);

        System.out.println("status=sucesso");
        System.out.println("valida=" + valida);
        System.out.println("mensagem=" + (valida ? "Assinatura válida." : "Assinatura inválida."));
    }
}
