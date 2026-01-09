// WebAuthn helper functions

const WebAuthn = {
    decode(base64) {
        const binary = atob(base64.replace(/-/g, '+').replace(/_/g, '/'));
        const bytes = new Uint8Array(binary.length);
        for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
        return bytes.buffer;
    },

    encode(buffer) {
        const bytes = new Uint8Array(buffer);
        let binary = '';
        for (let i = 0; i < bytes.byteLength; i++) binary += String.fromCharCode(bytes[i]);
        return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    },

    prepareCreate(options) {
        options.challenge = this.decode(options.challenge);
        options.user.id = this.decode(options.user.id);
        if (options.excludeCredentials) {
            options.excludeCredentials = options.excludeCredentials.map(c => ({
                ...c, id: this.decode(c.id)
            }));
        }
        return options;
    },

    prepareGet(options) {
        options.challenge = this.decode(options.challenge);
        if (options.allowCredentials) {
            options.allowCredentials = options.allowCredentials.map(c => ({
                ...c, id: this.decode(c.id)
            }));
        }
        return options;
    },

    formatCreateResponse(credential) {
        return {
            id: credential.id,
            rawId: this.encode(credential.rawId),
            type: credential.type,
            response: {
                clientDataJSON: this.encode(credential.response.clientDataJSON),
                attestationObject: this.encode(credential.response.attestationObject)
            }
        };
    },

    formatGetResponse(credential) {
        return {
            id: credential.id,
            rawId: this.encode(credential.rawId),
            type: credential.type,
            response: {
                clientDataJSON: this.encode(credential.response.clientDataJSON),
                authenticatorData: this.encode(credential.response.authenticatorData),
                signature: this.encode(credential.response.signature),
                userHandle: credential.response.userHandle ? this.encode(credential.response.userHandle) : null
            }
        };
    },

    async post(url, csrfToken, body = null) {
        const resp = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': csrfToken },
            body: body ? JSON.stringify(body) : undefined
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || 'Request failed');
        }
        return resp.json();
    }
};
