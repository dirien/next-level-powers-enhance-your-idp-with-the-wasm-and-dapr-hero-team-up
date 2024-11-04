# Specify the base image
FROM --platform=linux/amd64 node:23.1.0-bullseye-slim

# Set the working directory inside the container
WORKDIR /app

# Copy package.json and package-lock.json
COPY package*.json ./

# Install dependencies
RUN npm install

# Bundle app source inside Docker image
COPY . .

# Expose port and start application
EXPOSE 3000
CMD ["npm", "start"]
