package com.hubsaude.assinador.infrastructure.json;

import com.hubsaude.assinador.domain.SignatureResponse;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class JsonMapperTest {

    @Test
    void toJson_respostaValida_conteudoCorreto() {
        SignatureResponse r = new SignatureResponse("SIG==", true, "ok");
        String json = JsonMapper.toJson(r);
        assertTrue(json.contains("\"signature\":\"SIG==\""), "campo signature");
        assertTrue(json.contains("\"valid\":true"), "campo valid");
        assertTrue(json.contains("\"message\":\"ok\""), "campo message");
    }

    @Test
    void toJson_signatureNula_incluiNullSemAspas() {
        SignatureResponse r = new SignatureResponse(null, false, "erro");
        String json = JsonMapper.toJson(r);
        assertTrue(json.contains("\"signature\":null"), "signature null sem aspas");
        assertTrue(json.contains("\"valid\":false"), "campo valid false");
    }

    @Test
    void toJson_messageComAspas_escapaCorretamente() {
        SignatureResponse r = new SignatureResponse(null, false, "erro \"especial\"");
        String json = JsonMapper.toJson(r);
        assertTrue(json.contains("\\\"especial\\\""), "aspas escapadas no JSON");
    }
}
