package com.hubsaude.assinador;

public class AssinadorService {

    public static void sign(String[] args) {
        String conteudo = null;
        String algoritmo = "SHA256withRSA";

        for (int i = 1; i < args.length; i++) {
            switch (args[i]) {
                case "--content" -> {
                    if (i + 1 < args.length) conteudo = args[++i];
                }
                case "--algorithm" -> {
                    if (i + 1 < args.length) algoritmo = args[++i];
                }
            }
        }

        if (conteudo == null || conteudo.isBlank()) {
            System.err.println("Erro: parâmetro --content é obrigatório.");
            System.exit(1);
        }

        if (!algoritmo.equals("SHA256withRSA") && !algoritmo.equals("SHA512withRSA")) {
            System.err.println("Erro: algoritmo inválido: " + algoritmo);
            System.err.println("Algoritmos suportados: SHA256withRSA, SHA512withRSA");
            System.exit(1);
        }

        // Simulação de assinatura
        String assinaturaSimulada = "ASSINATURA-SIMULADA-" + algoritmo + "-" +
                Integer.toHexString(conteudo.hashCode()).toUpperCase();

        System.out.println("status=sucesso");
        System.out.println("assinatura=" + assinaturaSimulada);
        System.out.println("algoritmo=" + algoritmo);
    }

    public static void validate(String[] args) {
        String conteudo = null;
        String assinatura = null;

        for (int i = 1; i < args.length; i++) {
            switch (args[i]) {
                case "--content" -> {
                    if (i + 1 < args.length) conteudo = args[++i];
                }
                case "--signature" -> {
                    if (i + 1 < args.length) assinatura = args[++i];
                }
            }
        }

        if (conteudo == null || conteudo.isBlank()) {
            System.err.println("Erro: parâmetro --content é obrigatório.");
            System.exit(1);
        }

        if (assinatura == null || assinatura.isBlank()) {
            System.err.println("Erro: parâmetro --signature é obrigatório.");
            System.exit(1);
        }

        // Simulação de validação
        boolean valida = assinatura.startsWith("ASSINATURA-SIMULADA-");

        System.out.println("status=sucesso");
        System.out.println("valida=" + valida);
        System.out.println("mensagem=" + (valida ? "Assinatura válida." : "Assinatura inválida."));
    }
}