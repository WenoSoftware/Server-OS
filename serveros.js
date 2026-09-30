class ServerOSClient {
    constructor(port = 8080) {
        this.baseUrl = `http://127.0.0.1:${port}`;
    }

    // Check if the local Server OS node is active and healthy
    async getStatus() {
        try {
            const response = await fetch(`${this.baseUrl}/status`);
            if (!response.ok) throw new Error('Node unreachable');
            return await response.json();
        } catch (error) {
            console.error('[ServerOS SDK] Node offline or unresponsive:', error);
            return { online: false };
        }
    }

    // Toggle node state (e.g., pause/resume peer sharing)
    async toggleNode() {
        try {
            const response = await fetch(`${this.baseUrl}/toggle`, { method: 'POST' });
            return await response.json();
        } catch (error) {
            console.error('[ServerOS SDK] Failed to toggle node state:', error);
            throw error;
        }
    }

    // Programmatically push raw data or text to be chunked and stored locally
    async publishData(filename, content) {
        try {
            const formData = new FormData();
            const blob = new Blob([content], { type: 'text/html' });
            formData.append('file', blob, filename);

            const response = await fetch(`${this.baseUrl}/publish`, {
                method: 'POST',
                body: formData
            });

            if (!response.ok) throw new Error('Failed to publish data to local node');
            
            const result = await response.json();
            return result; // Expected to return { manifestHash: "..." }
        } catch (error) {
            console.error('[ServerOS SDK] Publishing error:', error);
            throw error;
        }
    }

    // Get the gateway URL for a specific P2P manifest hash
    getSiteUrl(manifestHash) {
        return `${this.baseUrl}/site/${manifestHash}`;
    }

    // Fetch raw chunk data directly from local storage or swarm
    async getChunk(hash) {
        try {
            const response = await fetch(`${this.baseUrl}/get-chunk?hash=${hash}`);
            if (!response.ok) throw new Error(`Chunk not found: ${hash}`);
            return await response.arrayBuffer();
        } catch (error) {
            console.error('[ServerOS SDK] Error fetching chunk:', error);
            throw error;
        }
    }
}

export default ServerOSClient;