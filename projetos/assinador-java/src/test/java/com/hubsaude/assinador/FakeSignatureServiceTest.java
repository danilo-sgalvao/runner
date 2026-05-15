package com.hubsaude.assinador;

import com.hubsaude.assinador.domain.SignRequest;
import com.hubsaude.assinador.domain.ValidateRequest;
import com.hubsaude.assinador.domain.SignatureResponse;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class FakeSignatureServiceTest {

    private final SignatureService service = new FakeSignatureService();

    // ------------------------------------------------------------------ sign

    @Test
    void sign_conteudoValido_retornaAssinaturaSimulada() {
        SignRequest request = new SignRequest();
        request.setContent("documento teste");

        SignatureResponse response = service.sign(request);

        assertNotNull(response);
        assertTrue(response.isValid());
        assertEquals(FakeSignatureService.FAKE_SIGNATURE, response.getSignature());
        assertEquals("Assinatura criada com sucesso", response.getMessage());
    }

    @Test
    void sign_conteudoNulo_retornaErro() {
        SignRequest request = new SignRequest();
        request.setContent(null);

        SignatureResponse response = service.sign(request);

        assertNotNull(response);
        assertFalse(response.isValid());
        assertNull(response.getSignature());
        assertTrue(response.getMessage().contains("content"));
    }

    @Test
    void sign_conteudoVazio_retornaErro() {
        SignRequest request = new SignRequest();
        request.setContent("");

        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
        assertNull(response.getSignature());
    }

    @Test
    void sign_conteudoApenasEspacos_retornaErro() {
        SignRequest request = new SignRequest();
        request.setContent("   ");

        SignatureResponse response = service.sign(request);

        assertFalse(response.isValid());
        assertNull(response.getSignature());
    }

    // --------------------------------------------------------------- validate

    @Test
    void validate_assinaturaCorreta_retornaValida() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        request.setSignature(FakeSignatureService.FAKE_SIGNATURE);

        SignatureResponse response = service.validate(request);

        assertNotNull(response);
        assertTrue(response.isValid());
        assertEquals("Assinatura é válida", response.getMessage());
    }

    @Test
    void validate_assinaturaErrada_retornaInvalida() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        request.setSignature("ASSINATURA-ERRADA");

        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertEquals("Assinatura é inválida", response.getMessage());
    }

    @Test
    void validate_conteudoNulo_retornaErro() {
        ValidateRequest request = new ValidateRequest();
        request.setContent(null);
        request.setSignature(FakeSignatureService.FAKE_SIGNATURE);

        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("content"));
    }

    @Test
    void validate_conteudoApenasEspacos_retornaErro() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("   ");
        request.setSignature(FakeSignatureService.FAKE_SIGNATURE);

        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
    }

    @Test
    void validate_assinaturaNula_retornaErro() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        request.setSignature(null);

        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
        assertTrue(response.getMessage().contains("signature"));
    }

    @Test
    void validate_assinaturaVazia_retornaErro() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        request.setSignature("");

        SignatureResponse response = service.validate(request);

        assertFalse(response.isValid());
    }

    // --------------------------------------------------------- fluxo completo

    @Test
    void fluxoCompleto_signEntaoValidate_retornaValida() {
        SignRequest signReq = new SignRequest();
        signReq.setContent("documento importante");
        SignatureResponse signed = service.sign(signReq);
        assertTrue(signed.isValid());

        ValidateRequest validateReq = new ValidateRequest();
        validateReq.setContent("documento importante");
        validateReq.setSignature(signed.getSignature());
        SignatureResponse validated = service.validate(validateReq);

        assertTrue(validated.isValid());
    }

    // ------------------------------------------------------ toJson (Main)

    @Test
    void toJson_respostaValida_produziJsonCorreto() {
        SignatureResponse r = new SignatureResponse("SIG==", true, "ok");
        String json = Main.toJson(r);
        assertEquals("{\"signature\":\"SIG==\",\"valid\":true,\"message\":\"ok\"}", json);
    }

    @Test
    void toJson_signatureNula_produziNullSemAspas() {
        SignatureResponse r = new SignatureResponse(null, false, "erro");
        String json = Main.toJson(r);
        assertEquals("{\"signature\":null,\"valid\":false,\"message\":\"erro\"}", json);
    }

    @Test
    void toJson_messageComAspas_escapaCorretamente() {
        SignatureResponse r = new SignatureResponse(null, false, "erro \"especial\"");
        String json = Main.toJson(r);
        assertTrue(json.contains("\\\"especial\\\""));
    }
}
