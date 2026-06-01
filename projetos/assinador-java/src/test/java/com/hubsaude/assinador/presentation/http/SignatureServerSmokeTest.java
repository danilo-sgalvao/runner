package com.hubsaude.assinador.presentation.http;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;

import java.util.Map;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Teste de fumaça do modo servidor: ao contrário de {@link SignatureControllerTest}
 * (que usa MockMvc e não abre porta), este sobe um Tomcat real em porta aleatória
 * (webEnvironment = RANDOM_PORT) e faz requisições HTTP de verdade pela rede local,
 * exercitando o caminho ponta-a-ponta servidor → controller → use case.
 *
 * As respostas são desserializadas em Map para depender apenas do JSON do contrato,
 * sem acoplar o teste ao DTO {@code SignatureHttpResponse} (que é imutável e não
 * tem construtor/setters para desserialização do lado cliente).
 */
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
class SignatureServerSmokeTest {

    @Autowired
    private TestRestTemplate rest;

    @Test
    void postSignRealHttp_retorna200ComAssinatura() {
        ResponseEntity<Map> resp = rest.postForEntity(
                "/sign", Map.of("content", "documento"), Map.class);

        assertEquals(HttpStatus.OK, resp.getStatusCode());
        assertEquals(true, resp.getBody().get("valid"));
        assertEquals("MOCKED_SIGNATURE_BASE64_==", resp.getBody().get("signature"));
    }

    @Test
    void postValidateRealHttp_assinaturaCorreta_retorna200Valid() {
        ResponseEntity<Map> resp = rest.postForEntity(
                "/validate",
                Map.of("content", "doc", "signature", "MOCKED_SIGNATURE_BASE64_=="),
                Map.class);

        assertEquals(HttpStatus.OK, resp.getStatusCode());
        assertTrue((Boolean) resp.getBody().get("valid"));
    }

    @Test
    void postSignRealHttp_contentVazio_retorna400() {
        ResponseEntity<Map> resp = rest.postForEntity(
                "/sign", Map.of("content", ""), Map.class);

        assertEquals(HttpStatus.BAD_REQUEST, resp.getStatusCode());
        assertEquals("Parâmetro 'content' inválido ou ausente", resp.getBody().get("message"));
    }
}
