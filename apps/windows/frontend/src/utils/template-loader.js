// Template loader utility
export class TemplateLoader {
    static async loadTemplate(templatePath) {
        try {
            const response = await fetch(templatePath);
            if (!response.ok) {
                throw new Error(`Failed to load template: ${response.statusText}`);
            }
            return await response.text();
        } catch (error) {
            console.error('Error loading template:', error);
            throw error;
        }
    }

    static renderTemplate(container, templateContent) {
        if (typeof container === 'string') {
            container = document.querySelector(container);
        }
        
        if (!container) {
            throw new Error('Container element not found');
        }
        
        container.innerHTML = templateContent;
    }
}
