package com.hubsaude.assinador;

public class AssinadorService {

    private static final String[] ALGORITMOS_SUPORTADOS = {"SHA256withRSA", "SHA512withRSA"};
    private static final String PREFIXO_ASSINATURA = "ASSINATURA-SIMULADA-";

    /**
     * Simula a criação de uma assinatura digital.
     *
     * @param conteudo  texto a ser assinado (obrigatório, não vazio)
     * @param algoritmo algoritmo de assinatura (SHA256withRSA ou SHA512withRSA)
     * @return string representando a assinatura simulada
     * @throws IllegalArgumentException se algum parâmetro for inválido
     */
    public static String sign(String conteudo, String algoritmo) {
        if (conteudo == null || conteudo.isBlank()) {
            throw new IllegalArgumentException("parâmetro --content é obrigatório e não pode ser vazio.");
        }

        if (algoritmo == null || algoritmo.isBlank()) {
            throw new IllegalArgumentException("parâmetro --algorithm é obrigatório.");
        }

        boolean algoritmoValido = false;
        for (String alg : ALGORITMOS_SUPORTADOS) {
            if (alg.equals(algoritmo)) {
                algoritmoValido = true;
                break;
            }
        }

        if (!algoritmoValido) {
            throw new IllegalArgumentException(
                "algoritmo inválido: '" + algoritmo + "'. " +
                "Algoritmos suportados: SHA256withRSA, SHA512withRSA"
            );
        }

        String hash = Integer.toHexString(conteudo.hashCode()).toUpperCase();
        return PREFIXO_ASSINATURA + algoritmo + "-" + hash;
    }

    /**
     * Simula a validação de uma assinatura digital.
     *
     * @param conteudo   texto original (obrigatório, não vazio)
     * @param assinatura assinatura a ser validada (obrigatório, não vazia)
     * @return true se a assinatura é considerada válida, false caso contrário
     * @throws IllegalArgumentException se algum parâmetro for inválido
     */
    public static boolean validate(String conteudo, String assinatura) {
        if (conteudo == null || conteudo.isBlank()) {
            throw new IllegalArgumentException("parâmetro --content é obrigatório e não pode ser vazio.");
        }

        if (assinatura == null || assinatura.isBlank()) {
            throw new IllegalArgumentException("parâmetro --signature é obrigatório e não pode ser vazio.");
        }

        return assinatura.startsWith(PREFIXO_ASSINATURA);
    }
}
