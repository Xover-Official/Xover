import * as vscode from 'vscode';
import { TalosAPIClient } from '../api/client';

export class AIConsolePanel {
    public static currentPanel: AIConsolePanel | undefined;
    private readonly _panel: vscode.WebviewPanel;
    private _disposables: vscode.Disposable[] = [];
    private _chatHistory: { role: string; content: string; }[] = [];

    private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri, private apiClient: TalosAPIClient) {
        this._panel = panel;
        this._update();

        this._panel.webview.onDidReceiveMessage(
            async message => {
                switch (message.command) {
                    case 'sendMessage':
                        await this.handleUserMessage(message.text);
                        break;
                }
            },
            null,
            this._disposables
        );

        this._panel.onDidDispose(() => this.dispose(), null, this._disposables);
    }

    public static render(context: vscode.ExtensionContext, apiClient: TalosAPIClient) {
        const column = vscode.window.activeTextEditor
            ? vscode.window.activeTextEditor.viewColumn
            : undefined;

        if (AIConsolePanel.currentPanel) {
            AIConsolePanel.currentPanel._panel.reveal(column);
            return;
        }

        const panel = vscode.window.createWebviewPanel(
            'talosAIConsole',
            'Talos AI Console',
            column || vscode.ViewColumn.Two,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        AIConsolePanel.currentPanel = new AIConsolePanel(panel, context.extensionUri, apiClient);
    }

    private async handleUserMessage(text: string) {
        this._chatHistory.push({ role: 'user', content: text });
        this._update();

        try {
            const response = await this.apiClient.chatWithAI(text);
            this._chatHistory.push({ role: 'assistant', content: response });
            this._update();
        } catch (error) {
            this._chatHistory.push({
                role: 'error',
                content: `Failed to get AI response: ${error}`
            });
            this._update();
        }
    }

    private _update() {
        this._panel.webview.html = this._getHtmlForWebview();
    }

    private _getHtmlForWebview(): string {
        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI Console</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            padding: 0;
            margin: 0;
            background: #1e1e2e;
            color: #fff;
            display: flex;
            flex-direction: column;
            height: 100vh;
        }
        .header {
            padding: 20px;
            background: linear-gradient(90deg, #60a5fa 0%, #a78bfa 100%);
            font-size: 18px;
            font-weight: 600;
        }
        .chat-container {
            flex: 1;
            overflow-y: auto;
            padding: 20px;
            display: flex;
            flex-direction: column;
            gap: 12px;
        }
        .message {
            padding: 12px 16px;
            border-radius: 12px;
            max-width: 80%;
        }
        .message.user {
            background: rgba(96, 165, 250, 0.2);
            border: 1px solid rgba(96, 165, 250, 0.4);
            align-self: flex-end;
        }
        .message.assistant {
            background: rgba(167, 139, 250, 0.2);
            border: 1px solid rgba(167, 139, 250, 0.4);
            align-self: flex-start;
        }
        .message.error {
            background: rgba(239, 68, 68, 0.2);
            border: 1px solid rgba(239, 68, 68, 0.4);
            align-self: center;
        }
        .input-container {
            padding: 20px;
            background: rgba(30, 30, 60, 0.8);
            border-top: 1px solid rgba(255, 255, 255, 0.1);
            display: flex;
            gap: 12px;
        }
        .input-field {
            flex: 1;
            padding: 12px 16px;
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 8px;
            color: #fff;
            font-size: 14px;
        }
        .send-button {
            padding: 12px 24px;
            background: linear-gradient(90deg, #60a5fa 0%, #a78bfa 100%);
            border: none;
            border-radius: 8px;
            color: white;
            font-weight: 600;
            cursor: pointer;
        }
        .send-button:hover {
            opacity: 0.9;
        }
    </style>
</head>
<body>
    <div class="header">💬 AI Console</div>
    <div class="chat-container" id="chatContainer">
        ${this._chatHistory.map(msg => `
            <div class="message ${msg.role}">
                ${msg.content}
            </div>
        `).join('')}
    </div>
    <div class="input-container">
        <input type="text" class="input-field" id="messageInput" placeholder="Ask the AI swarm anything..." />
        <button class="send-button" onclick="sendMessage()">Send</button>
    </div>

    <script>
        const vscode = acquireVsCodeApi();
        
        function sendMessage() {
            const input = document.getElementById('messageInput');
            const text = input.value.trim();
            
            if (text) {
                vscode.postMessage({ command: 'sendMessage', text });
                input.value = '';
            }
        }

        // Send on Enter
        document.getElementById('messageInput').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                sendMessage();
            }
        });

        // Auto-scroll to bottom
        const container = document.getElementById('chatContainer');
        container.scrollTop = container.scrollHeight;
    </script>
</body>
</html>`;
    }

    public dispose() {
        AIConsolePanel.currentPanel = undefined;
        this._panel.dispose();

        while (this._disposables.length) {
            const disposable = this._disposables.pop();
            if (disposable) {
                disposable.dispose();
            }
        }
    }
}
