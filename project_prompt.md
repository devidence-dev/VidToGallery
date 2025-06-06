
```markdown
# Video Saver Web App (iOS Focused) Named : VidToGallery
**Stack**: Go (Fiber) + Vuejs (vant) + Raspberry Pi ARM64

## **Requisitos Técnicos**
1. **Arquitectura Modular**
   ```
   /video-saver
   ├── backend/
   │   ├── cmd/
   │   ├── pkg/
   │   │   ├── downloader/  # Lógica específica por plataforma
   │   │   ├── api/         # Handlers HTTP
   │   │   └── config/      # Gestión de entorno
   │   └── go.mod
   ├── frontend/
   │   ├── src/
   │   │   ├── lib/
   │   │   ├── stores/      # vuejs stores reactivos
   │   │   └── components/  # UI optimizada para iOS
   │   └── package.json
   └── infra/
       ├── caddy/
       └── systemd/
   ```

2. **Flujo de Trabajo Principal**
   ```
   sequenceDiagram
   Usuario->>Frontend: Pega URL
   Frontend->>Backend: POST /process {url: "tiktok.com/..."}
   Backend->>Plataforma: Resuelve URL real
   Backend->>Backend: Descarga video (yt-dlp)
   Backend-->>Frontend: {video_url: "https://cdn..."}
   Frontend->>iOS: navigator.share(videoFile)
   iOS-->>Usuario: Video en carrete
   ```

3. **Implementación Backend (Go)**
   - **Requisitos**:
     - Descarga concurrente usando workers (goroutines)
     - Cacheo de videos con Redis (TTL 24h)
     - Rotación automática de User-Agents
   ```
   // Ejemplo handler de descarga
   func TikTokHandler(c *fiber.Ctx) error {
       url := c.Query("url")
       quality := c.Query("quality", "720")
       
       video, err := downloader.FetchTikTok(url, quality)
       if err != nil {
           return c.Status(500).JSON(fiber.Map{"error": err.Error()})
       }
       
       return c.Type("mp4").Send(video.Data)
   }
   ```

4. **Frontend iOS-Optimizado (vuejs)**
   - **Componentes Esenciales**:
     ```
     // VideoInput.vuejs
     
       let url = '';
       export let onProcess = () => {};
     
     
     
       
       Descargar
     
     ```

5. **Despliegue Raspberry Pi**
   - **Automación Requerida**:
     ```
     #! /bin/bash
     # build-arm.sh
     GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
       go build -ldflags="-s -w" -o video-saver-arm64
     ```

## **Requerimientos Específicos del Agente**
1. **Configuración Cross-Platform**
   - Generar `Dockerfile` multi-arch (amd64 + arm64)
   - Auto-detectar entorno Raspberry Pi en runtime

2. **Manejo de Formatos iOS**
   - Conversión automática a MOV/H265 si es necesario
   - Metadata EXIF con origen del video

3. **Optimización para Safari**
   - Polyfills para `HTML5 Video` en iOS 15+
   - Soporte para gestos táctiles (swipe to cancel)

4. **Seguridad**
   - Validación de URLs con regex por plataforma
   ```
   // Instagram URL validation
   var instagramRegex = regexp.MustCompile(`^(?:https?:\/\/)?(?:www\.)?instagram\.com\/.*\/?`)
   ```

5. **Monitoreo**
   - Integrar endpoint `/metrics` compatible con Prometheus
   - Trackear: 
     - Tiempos de descarga por plataforma
     - Resoluciones más solicitadas
     - Errores por política de contenido

## **Entregables Esperados**
- [ ] Docker-compose.yml con servicios: app, redis, caddy
- [ ] Script de post-instalación para Raspberry Pi OS
- [ ] Componente vuejs con preview del video antes de descargar
- [ ] API documentation en formato OpenAPI 3.0
- [ ] Sistema de logging unificado (FluentBit)
```

---

**Consideraciones Clave Extraídas de los Search Results**:
1. Usar `navigator.share()` para iOS Photo Library [4]
2. Implementar auto-rotación de User-Agents similar a yt-dlp [1][2]
3. Optimizar builds vuejs para <50KB (MDN guidelines [3])
4. Incluir sistema de colas para evitar bloqueos en Raspberry Pi [5]
