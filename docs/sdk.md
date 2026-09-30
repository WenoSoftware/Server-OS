# How to embed Weno Server OS into your websites and web applications

Server OS has a built-in JavaScript SDK that makes it surprisingly easy to use Server OS with your apps. All you need to do is paste this code into into your project's index.html file:

```html
<script type="module">
    import ServerOSClient from './serveros.js';

    const serverOS = new ServerOSClient(8080);

    async function initApp() {
        const status = await serverOS.getStatus();
        
        if (status.online !== false) {
            document.getElementById('node-status').innerText = "⚡ Local Server OS Node Active";
            
            const siteHash = "93a577558f9d80e5a0ab95efee1ed271eca6994305a955b2e6c642b75778eabb";
            console.log("Loading P2P site from:", serverOS.getSiteUrl(siteHash));
        } else {
            document.getElementById('node-status').innerText = "⚠️ Server OS Node Offline.";
        }
    }

    initApp();
</script>
```