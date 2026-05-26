package com.hubsaude.assinador.infrastructure.json;

import com.hubsaude.assinador.domain.model.SignatureResult;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class JsonMapperTest {

    @Test
    void toJson_respostaValida_conteudoCorreto() {
        SignatureResult r = new SignatureResult("SIG==", true, "ok");
        String json = JsonMapper.toJson(r);
        assertTrue(json.contains("\"signature\":\"SIG==\""), "campo signature");
        assertTrue(json.contains("\"valid\":true"), "campo valid");
        assertTrue(json.contains("\"message\":\"ok\""), "campo message");
    }

    @Test
    void toJson_signatureNula_incluiNullSemAspas() {
        SignatureResult r = new SignatureResult(null, false, "erro");
        String json = JsonMapper.toJson(r);
        assertTrue(json.contains("\"signature\":null"), "signature null sem aspas");
        assertTrue(json.contains("\"valid\":false"), "campo valid false");
    }

    @Test
    void toJson_messageComAspas_escapaCorretamente() {
        SignatureResult r = new SignatureResult(null, false, "erro \"especial\"");
        String json = JsonMapper.toJson(r);
        assertTrue(json.contains("\\\"especial\\\""), "aspas escapadas no JSON");
    }
}
