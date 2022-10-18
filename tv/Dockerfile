# Specify a base image
FROM node:latest as build-deps

EXPOSE 80

# Create working directory and copy the app before running yarn install as the artifactory
# credentials can be inside .npmrc
WORKDIR /usr/src/app
COPY . ./

# Run yarn install
RUN yarn install

# Build the project
CMD ["yarn", "run", "build"]

# Install serve command for yarn package manager
RUN yarn global add serve

# Navigate to build folder
WORKDIR /usr/src/app/build

# Start the application
CMD serve -p 80 -s . 