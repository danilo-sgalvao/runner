package com.hubsaude.assinador;

public class Main {
    public static void main(String[] args) {
        if (args.length == 0) {
            System.err.println("Erro: nenhum comando fornecido.");
            System.err.println("Uso: assinador <comando> [opções]");
            System.exit(1);
        }

        String comando = args[0];

        switch (comando) {
            case "sign" -> AssinadorService.sign(args);
            case "validate" -> AssinadorService.validate(args);
            default -> {
                System.err.println("Erro: comando desconhecido: " + comando);
                System.err.println("Comandos disponíveis: sign, validate");
                System.exit(1);
            }
        }
    }
}